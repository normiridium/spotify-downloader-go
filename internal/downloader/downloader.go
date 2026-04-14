package downloader

import (
    "bufio"
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

type Downloader struct {
    ProxySocksHost string
    AudioFormat    string
    AudioQuality   string
}

func (d Downloader) DownloadAudio(ctx context.Context, query string) (string, error) {
    if strings.TrimSpace(query) == "" {
        return "", errors.New("empty query")
    }

    tempDir, err := os.MkdirTemp("", "spotify-dl-*")
    if err != nil {
        return "", err
    }

    outputTemplate := filepath.Join(tempDir, "%(title)s.%(ext)s")

    args := []string{
        "-f", "bestaudio[ext=m4a]/bestaudio/best",
        "-x",
        "--audio-format", d.audioFormatOrDefault(),
        "--audio-quality", d.audioQualityOrDefault(),
        "--no-playlist",
        "--extractor-args", "youtube:player_client=android,web",
        "--print", "after_move:filepath",
        "-o", outputTemplate,
        fmt.Sprintf("ytsearch1:%s", query),
    }

    if d.ProxySocksHost != "" {
        args = append([]string{"--proxy", "socks5://" + d.ProxySocksHost}, args...)
    }

    ytdlpBin := strings.TrimSpace(os.Getenv("YTDLP_BIN"))
    if ytdlpBin == "" {
        ytdlpBin = "yt-dlp"
    }
    cmd := exec.CommandContext(ctx, ytdlpBin, args...)
    cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return "", err
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return "", err
    }

    if err := cmd.Start(); err != nil {
        return "", err
    }

    var lastPath string
    scan := bufio.NewScanner(stdout)
    for scan.Scan() {
        line := strings.TrimSpace(scan.Text())
        if line != "" {
            lastPath = line
        }
    }

    errBuf := new(strings.Builder)
    errScan := bufio.NewScanner(stderr)
    for errScan.Scan() {
        errBuf.WriteString(errScan.Text())
        errBuf.WriteString("\n")
    }

    if err := cmd.Wait(); err != nil {
        return "", fmt.Errorf("yt-dlp failed: %w: %s", err, strings.TrimSpace(errBuf.String()))
    }

    if lastPath == "" {
        return "", errors.New("yt-dlp returned no output file path")
    }

    if _, err := os.Stat(lastPath); err != nil {
        return "", fmt.Errorf("downloaded file not found: %w", err)
    }

    return lastPath, nil
}

func DefaultContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), 5*time.Minute)
}

func (d Downloader) audioFormatOrDefault() string {
    if strings.TrimSpace(d.AudioFormat) == "" {
        return "mp3"
    }
    return strings.TrimSpace(d.AudioFormat)
}

func (d Downloader) audioQualityOrDefault() string {
    if strings.TrimSpace(d.AudioQuality) == "" {
        return "320K"
    }
    return strings.TrimSpace(d.AudioQuality)
}
