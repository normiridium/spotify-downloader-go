package config

import (
    "errors"
    "os"
    "strconv"
    "strings"
)

type Config struct {
    BotToken            string
    SpotifyClientID     string
    SpotifyClientSecret string
    OwnerID             int64
    AllowedChats        map[int64]struct{}
    ProxySocksHost      string
    AudioFormat         string
    AudioQuality        string
}

func Load() (Config, error) {
    cfg := Config{}

    cfg.BotToken = strings.TrimSpace(os.Getenv("BOT_TOKEN"))
    if cfg.BotToken == "" {
        return cfg, errors.New("BOT_TOKEN is required")
    }

    cfg.SpotifyClientID = firstNonEmpty(
        os.Getenv("SPOTIPY_CLIENT_ID"),
        os.Getenv("SPOTIFY_CLIENT_ID"),
    )
    cfg.SpotifyClientSecret = firstNonEmpty(
        os.Getenv("SPOTIPY_CLIENT_SECRET"),
        os.Getenv("SPOTIFY_CLIENT_SECRET"),
    )
    if cfg.SpotifyClientID == "" || cfg.SpotifyClientSecret == "" {
        return cfg, errors.New("SPOTIPY_CLIENT_ID and SPOTIPY_CLIENT_SECRET are required")
    }

    if owner := strings.TrimSpace(os.Getenv("OWNER_ID")); owner != "" {
        if v, err := strconv.ParseInt(owner, 10, 64); err == nil {
            cfg.OwnerID = v
        }
    }

    cfg.AllowedChats = parseChatIDs(os.Getenv("AUTH_CHATS"))
    cfg.ProxySocksHost = strings.TrimSpace(os.Getenv("FIXIE_SOCKS_HOST"))
    cfg.AudioFormat = firstNonEmpty(os.Getenv("AUDIO_FORMAT"), "mp3")
    cfg.AudioQuality = firstNonEmpty(os.Getenv("AUDIO_QUALITY"), "320K")

    return cfg, nil
}

func parseChatIDs(raw string) map[int64]struct{} {
    out := make(map[int64]struct{})
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return out
    }
    parts := strings.Fields(raw)
    for _, p := range parts {
        if id, err := strconv.ParseInt(p, 10, 64); err == nil {
            out[id] = struct{}{}
        }
    }
    return out
}

func firstNonEmpty(values ...string) string {
    for _, v := range values {
        if strings.TrimSpace(v) != "" {
            return strings.TrimSpace(v)
        }
    }
    return ""
}
