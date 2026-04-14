package main

import (
    "strings"
    "testing"

    "spotify-downloader-go/internal/spotify"
)

func TestJoinArtists(t *testing.T) {
    artists := []spotify.Artist{{Name: "A"}, {Name: "B"}}
    got := joinArtists(artists)
    if got != "A, B" {
        t.Fatalf("unexpected joinArtists: %s", got)
    }
}

func TestBuildQuery(t *testing.T) {
    track := spotify.Track{Name: "Song", Artists: []spotify.Artist{{Name: "Artist"}}}
    got := buildQuery(track)
    if got != "Song - Artist" {
        t.Fatalf("unexpected buildQuery: %s", got)
    }

    track = spotify.Track{Name: "", Artists: []spotify.Artist{{Name: "Artist"}}}
    if buildQuery(track) != "Artist" {
        t.Fatalf("expected artist only")
    }
}

func TestFormatCaption(t *testing.T) {
    track := spotify.Track{
        Name: "T[st]",
        Artists: []spotify.Artist{{Name: "Ar_t"}},
        Album: spotify.Album{Name: "Al(b)", ReleaseDate: "2020-01-01"},
    }

    caption := formatCaption(track)
    if caption == "" {
        t.Fatal("caption should not be empty")
    }
    if !containsAll(caption, []string{"\\[", "\\]", "\\_", "\\(", "\\)"}) {
        t.Fatalf("expected markdown escaping, got: %s", caption)
    }
}

func containsAll(s string, subs []string) bool {
    for _, sub := range subs {
        if !strings.Contains(s, sub) {
            return false
        }
    }
    return true
}
