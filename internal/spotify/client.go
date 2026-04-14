package spotify

import (
    "bytes"
    "context"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "time"
)

type Client struct {
    httpClient *http.Client
    clientID   string
    secret     string

    mu        sync.Mutex
    token     string
    expiresAt time.Time
}

func New(clientID, secret string) *Client {
    return &Client{
        httpClient: &http.Client{Timeout: 20 * time.Second},
        clientID:   clientID,
        secret:     secret,
    }
}

func (c *Client) ensureToken(ctx context.Context) (string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.token != "" && time.Until(c.expiresAt) > 30*time.Second {
        return c.token, nil
    }

    form := url.Values{}
    form.Set("grant_type", "client_credentials")

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(form.Encode()))
    if err != nil {
        return "", err
    }

    auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.secret))
    req.Header.Set("Authorization", "Basic "+auth)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("spotify token request failed: %s", string(bytes.TrimSpace(body)))
    }

    var payload struct {
        AccessToken string `json:"access_token"`
        ExpiresIn   int    `json:"expires_in"`
        TokenType   string `json:"token_type"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
        return "", err
    }
    if payload.AccessToken == "" {
        return "", errors.New("spotify token response missing access_token")
    }

    c.token = payload.AccessToken
    c.expiresAt = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)

    return c.token, nil
}

func (c *Client) do(ctx context.Context, method, endpoint string, out any) error {
    token, err := c.ensureToken(ctx)
    if err != nil {
        return err
    }

    req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("spotify api error: %s", string(bytes.TrimSpace(body)))
    }

    return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) GetTrack(ctx context.Context, id string) (Track, error) {
    var t Track
    err := c.do(ctx, http.MethodGet, "https://api.spotify.com/v1/tracks/"+id, &t)
    return t, err
}

func (c *Client) GetAlbum(ctx context.Context, id string) (Album, error) {
    var a Album
    err := c.do(ctx, http.MethodGet, "https://api.spotify.com/v1/albums/"+id, &a)
    return a, err
}

func (c *Client) GetPlaylist(ctx context.Context, id string, limit, offset int) (PlaylistTracksPage, error) {
    var p PlaylistTracksPage
    endpoint := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?limit=%d&offset=%d", id, limit, offset)
    err := c.do(ctx, http.MethodGet, endpoint, &p)
    return p, err
}

func (c *Client) GetAlbumTracks(ctx context.Context, id string, limit, offset int) (AlbumTracksPage, error) {
    var p AlbumTracksPage
    endpoint := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=%d&offset=%d", id, limit, offset)
    err := c.do(ctx, http.MethodGet, endpoint, &p)
    return p, err
}
