package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

    "spotify-downloader-go/internal/config"
    "spotify-downloader-go/internal/downloader"
    "spotify-downloader-go/internal/spotify"
    "spotify-downloader-go/internal/util"
)

const helpText = "" +
    "Send me a Spotify link and I will fetch the audio from YouTube.\n" +
    "\n" +
    "Commands:\n" +
    "/start - show this message\n" +
    "/help - show this message\n" +
    "/thumb <link> - send cover image\n" +
    "/preview <link> - send Spotify preview clip if available\n"

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
    if err != nil {
        log.Fatal(err)
    }

    bot.Debug = false
    log.Printf("authorized on account %s", bot.Self.UserName)

    sp := spotify.New(cfg.SpotifyClientID, cfg.SpotifyClientSecret)
    dl := downloader.Downloader{
        ProxySocksHost: cfg.ProxySocksHost,
        AudioFormat:    cfg.AudioFormat,
        AudioQuality:   cfg.AudioQuality,
    }

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 30

    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil {
            continue
        }

        msg := update.Message
        if !isAllowed(cfg.AllowedChats, msg.Chat.ID) {
            continue
        }

        if msg.IsCommand() {
            switch msg.Command() {
            case "start", "help":
                reply(bot, msg.Chat.ID, helpText)
            case "thumb":
                handleThumb(bot, sp, msg)
            case "preview":
                handlePreview(bot, sp, msg)
            default:
                reply(bot, msg.Chat.ID, "Unknown command. Use /help.")
            }
            continue
        }

        url := extractSpotifyURL(msg.Text)
        if url == "" {
            continue
        }

        handleDownload(bot, sp, dl, msg, url)
    }
}

func isAllowed(allowed map[int64]struct{}, chatID int64) bool {
    if len(allowed) == 0 {
        return true
    }
    _, ok := allowed[chatID]
    return ok
}

func reply(bot *tgbotapi.BotAPI, chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.DisableWebPagePreview = true
    _, _ = bot.Send(msg)
}

func handleThumb(bot *tgbotapi.BotAPI, sp *spotify.Client, msg *tgbotapi.Message) {
    url := extractSpotifyURL(msg.CommandArguments())
    if url == "" {
        reply(bot, msg.Chat.ID, "Send /thumb <spotify link>")
        return
    }

    parsed, err := util.ParseSpotifyURL(url)
    if err != nil {
        reply(bot, msg.Chat.ID, "Could not parse Spotify link")
        return
    }

    if parsed.Type != "track" && parsed.Type != "album" && parsed.Type != "playlist" {
        reply(bot, msg.Chat.ID, "Unsupported Spotify type")
        return
    }

    if parsed.Type == "track" {
        track, err := sp.GetTrack(context.Background(), parsed.ID)
        if err != nil {
            reply(bot, msg.Chat.ID, "Spotify error: "+err.Error())
            return
        }
        if len(track.Album.Images) == 0 {
            reply(bot, msg.Chat.ID, "No cover image")
            return
        }
        photo := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FileURL(track.Album.Images[0].URL))
        _, _ = bot.Send(photo)
        return
    }

    if parsed.Type == "album" {
        album, err := sp.GetAlbum(context.Background(), parsed.ID)
        if err != nil {
            reply(bot, msg.Chat.ID, "Spotify error: "+err.Error())
            return
        }
        if len(album.Images) == 0 {
            reply(bot, msg.Chat.ID, "No cover image")
            return
        }
        photo := tgbotapi.NewPhoto(msg.Chat.ID, tgbotapi.FileURL(album.Images[0].URL))
        _, _ = bot.Send(photo)
        return
    }

    reply(bot, msg.Chat.ID, "Playlist thumbnails are not implemented yet")
}

func handlePreview(bot *tgbotapi.BotAPI, sp *spotify.Client, msg *tgbotapi.Message) {
    url := extractSpotifyURL(msg.CommandArguments())
    if url == "" {
        reply(bot, msg.Chat.ID, "Send /preview <spotify link>")
        return
    }

    parsed, err := util.ParseSpotifyURL(url)
    if err != nil {
        reply(bot, msg.Chat.ID, "Could not parse Spotify link")
        return
    }
    if parsed.Type != "track" {
        reply(bot, msg.Chat.ID, "Preview is only supported for tracks")
        return
    }

    track, err := sp.GetTrack(context.Background(), parsed.ID)
    if err != nil {
        reply(bot, msg.Chat.ID, "Spotify error: "+err.Error())
        return
    }
    if track.PreviewURL == "" {
        reply(bot, msg.Chat.ID, "No preview available for this track")
        return
    }

    audio := tgbotapi.NewAudio(msg.Chat.ID, tgbotapi.FileURL(track.PreviewURL))
    audio.Title = track.Name
    audio.Performer = joinArtists(track.Artists)
    _, _ = bot.Send(audio)
}

func handleDownload(bot *tgbotapi.BotAPI, sp *spotify.Client, dl downloader.Downloader, msg *tgbotapi.Message, url string) {
    parsed, err := util.ParseSpotifyURL(url)
    if err != nil {
        reply(bot, msg.Chat.ID, "Could not parse Spotify link")
        return
    }

    switch parsed.Type {
    case "track":
        if err := downloadAndSendTrack(bot, sp, dl, msg.Chat.ID, parsed.ID); err != nil {
            reply(bot, msg.Chat.ID, "Error: "+err.Error())
        }
    case "album":
        if err := downloadAndSendAlbum(bot, sp, dl, msg.Chat.ID, parsed.ID); err != nil {
            reply(bot, msg.Chat.ID, "Error: "+err.Error())
        }
    case "playlist":
        if err := downloadAndSendPlaylist(bot, sp, dl, msg.Chat.ID, parsed.ID); err != nil {
            reply(bot, msg.Chat.ID, "Error: "+err.Error())
        }
    default:
        reply(bot, msg.Chat.ID, "Unsupported Spotify link type")
    }
}

func downloadAndSendTrack(bot *tgbotapi.BotAPI, sp *spotify.Client, dl downloader.Downloader, chatID int64, trackID string) error {
    track, err := sp.GetTrack(contextOrBackground(), trackID)
    if err != nil {
        return err
    }

    query := buildQuery(track)
    ctx, cancel := downloader.DefaultContext()
    defer cancel()

    path, err := dl.DownloadAudio(ctx, query)
    if err != nil {
        return err
    }
    defer os.Remove(path)

    audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(path))
    audio.Title = track.Name
    audio.Performer = joinArtists(track.Artists)
    audio.Caption = formatCaption(track)
    audio.ParseMode = "MarkdownV2"
    _, err = bot.Send(audio)
    return err
}

func buildQuery(track spotify.Track) string {
    title := strings.TrimSpace(track.Name)
    artist := strings.TrimSpace(joinArtists(track.Artists))
    if title == "" && artist == "" {
        return ""
    }
    if title == "" {
        return artist
    }
    if artist == "" {
        return title
    }
    return fmt.Sprintf("%s - %s", title, artist)
}

func downloadAndSendAlbum(bot *tgbotapi.BotAPI, sp *spotify.Client, dl downloader.Downloader, chatID int64, albumID string) error {
    album, err := sp.GetAlbum(contextOrBackground(), albumID)
    if err != nil {
        return err
    }

    total := album.TotalTracks
    offset := 0
    limit := 50

    for offset < total {
        page, err := sp.GetAlbumTracks(contextOrBackground(), albumID, limit, offset)
        if err != nil {
            return err
        }

        for _, track := range page.Items {
            track.Album = album
            if err := downloadAndSendTrack(bot, sp, dl, chatID, track.ID); err != nil {
                return err
            }
            time.Sleep(1 * time.Second)
        }

        offset += limit
        if len(page.Items) == 0 {
            break
        }
    }

    return nil
}

func downloadAndSendPlaylist(bot *tgbotapi.BotAPI, sp *spotify.Client, dl downloader.Downloader, chatID int64, playlistID string) error {
    offset := 0
    limit := 50
    for {
        page, err := sp.GetPlaylist(contextOrBackground(), playlistID, limit, offset)
        if err != nil {
            return err
        }

        for _, item := range page.Items {
            if item.Track.ID == "" {
                continue
            }
            if err := downloadAndSendTrack(bot, sp, dl, chatID, item.Track.ID); err != nil {
                return err
            }
            time.Sleep(1 * time.Second)
        }

        offset += limit
        if offset >= page.Total || len(page.Items) == 0 {
            break
        }
    }

    return nil
}

func formatCaption(track spotify.Track) string {
    year := ""
    if track.Album.ReleaseDate != "" {
        year = track.Album.ReleaseDate
        if len(year) > 4 {
            year = year[:4]
        }
    }

    album := track.Album.Name
    artists := joinArtists(track.Artists)

    lines := []string{
        fmt.Sprintf("*Title:* %s", escapeMD(track.Name)),
        fmt.Sprintf("*Artist:* %s", escapeMD(artists)),
    }
    if album != "" {
        lines = append(lines, fmt.Sprintf("*Album:* %s", escapeMD(album)))
    }
    if year != "" {
        lines = append(lines, fmt.Sprintf("*Year:* %s", escapeMD(year)))
    }
    return strings.Join(lines, "\n")
}

func joinArtists(artists []spotify.Artist) string {
    names := make([]string, 0, len(artists))
    for _, a := range artists {
        if strings.TrimSpace(a.Name) != "" {
            names = append(names, a.Name)
        }
    }
    return strings.Join(names, ", ")
}

func escapeMD(s string) string {
    replacer := strings.NewReplacer(
        "_", "\\_",
        "*", "\\*",
        "[", "\\[",
        "]", "\\]",
        "(", "\\(",
        ")", "\\)",
        "~", "\\~",
        "`", "\\`",
        ">", "\\>",
        "#", "\\#",
        "+", "\\+",
        "-", "\\-",
        "=", "\\=",
        "|", "\\|",
        "{", "\\{",
        "}", "\\}",
        ".", "\\.",
        "!", "\\!",
    )
    return replacer.Replace(s)
}

func extractSpotifyURL(text string) string {
    fields := strings.Fields(text)
    for _, f := range fields {
        if strings.Contains(f, "open.spotify.com") || strings.Contains(f, "spotify.link") {
            return strings.TrimSpace(f)
        }
    }
    return ""
}

func contextOrBackground() context.Context {
    return context.Background()
}
