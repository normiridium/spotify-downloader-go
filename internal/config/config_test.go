package config

import (
    "os"
    "testing"
)

func TestLoadConfig(t *testing.T) {
    t.Setenv("BOT_TOKEN", "test")
    t.Setenv("SPOTIPY_CLIENT_ID", "id")
    t.Setenv("SPOTIPY_CLIENT_SECRET", "secret")
    t.Setenv("OWNER_ID", "123")
    t.Setenv("AUTH_CHATS", "-1001 42")
    t.Setenv("AUDIO_FORMAT", "flac")
    t.Setenv("AUDIO_QUALITY", "0")

    cfg, err := Load()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if cfg.BotToken != "test" {
        t.Fatalf("unexpected bot token: %s", cfg.BotToken)
    }
    if cfg.SpotifyClientID != "id" || cfg.SpotifyClientSecret != "secret" {
        t.Fatalf("unexpected spotify creds")
    }
    if cfg.OwnerID != 123 {
        t.Fatalf("unexpected owner id: %d", cfg.OwnerID)
    }
    if len(cfg.AllowedChats) != 2 {
        t.Fatalf("unexpected allowed chats: %v", cfg.AllowedChats)
    }
    if cfg.AudioFormat != "flac" || cfg.AudioQuality != "0" {
        t.Fatalf("unexpected audio config: %s/%s", cfg.AudioFormat, cfg.AudioQuality)
    }
}

func TestLoadConfigMissing(t *testing.T) {
    os.Clearenv()
    if _, err := Load(); err == nil {
        t.Fatal("expected error when required vars missing")
    }
}

func TestLoadConfigAudioDefaults(t *testing.T) {
    t.Setenv("BOT_TOKEN", "test")
    t.Setenv("SPOTIPY_CLIENT_ID", "id")
    t.Setenv("SPOTIPY_CLIENT_SECRET", "secret")

    cfg, err := Load()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if cfg.AudioFormat != "mp3" || cfg.AudioQuality != "320K" {
        t.Fatalf("unexpected default audio config: %s/%s", cfg.AudioFormat, cfg.AudioQuality)
    }
}
