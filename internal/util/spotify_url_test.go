package util

import "testing"

func TestParseSpotifyURL_URI(t *testing.T) {
    got, err := ParseSpotifyURL("spotify:track:2QZ7WLBE8h2y1Y5Fb8RYbH")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got.Type != "track" || got.ID != "2QZ7WLBE8h2y1Y5Fb8RYbH" {
        t.Fatalf("unexpected result: %+v", got)
    }
}

func TestParseSpotifyURL_Open(t *testing.T) {
    got, err := ParseSpotifyURL("https://open.spotify.com/album/1ATL5GLyefJaxhQzSPVrLX")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got.Type != "album" || got.ID != "1ATL5GLyefJaxhQzSPVrLX" {
        t.Fatalf("unexpected result: %+v", got)
    }
}

func TestParseSpotifyURL_Invalid(t *testing.T) {
    if _, err := ParseSpotifyURL("https://example.com/track/123"); err == nil {
        t.Fatal("expected error for non-spotify URL")
    }
}
