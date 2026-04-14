package util

import (
    "errors"
    "net/http"
    "net/url"
    "strings"
    "time"
)

type ParsedSpotify struct {
    Type string
    ID   string
}

func ParseSpotifyURL(raw string) (ParsedSpotify, error) {
    if strings.TrimSpace(raw) == "" {
        return ParsedSpotify{}, errors.New("empty url")
    }

    if strings.HasPrefix(raw, "spotify:") {
        parts := strings.Split(raw, ":")
        if len(parts) >= 3 {
            return ParsedSpotify{Type: parts[1], ID: parts[2]}, nil
        }
        return ParsedSpotify{}, errors.New("invalid spotify uri")
    }

    // If it's already an open.spotify.com URL, parse directly to avoid network calls.
    if strings.Contains(raw, "open.spotify.com/") {
        return parseOpenSpotifyURL(raw)
    }

    resolved, err := resolveURL(raw)
    if err != nil {
        return ParsedSpotify{}, err
    }

    return parseOpenSpotifyURL(resolved)
}

func parseOpenSpotifyURL(raw string) (ParsedSpotify, error) {
    u, err := url.Parse(raw)
    if err != nil {
        return ParsedSpotify{}, err
    }

    if u.Host != "open.spotify.com" {
        return ParsedSpotify{}, errors.New("not an open.spotify.com url")
    }

    parts := strings.Split(strings.Trim(u.Path, "/"), "/")
    if len(parts) < 2 {
        return ParsedSpotify{}, errors.New("invalid spotify url path")
    }

    return ParsedSpotify{Type: parts[0], ID: parts[1]}, nil
}

func resolveURL(raw string) (string, error) {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    resp, err := client.Get(raw)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    return resp.Request.URL.String(), nil
}
