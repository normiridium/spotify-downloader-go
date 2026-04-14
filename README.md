# Spotify Downloader Bot (Go)

🎵 Telegram-бот для скачивания треков по Spotify-ссылкам.

Порт проекта: https://github.com/Masterolic/Spotify-Downloader

## ✨ Возможности
- 🤖 Telegram Bot API (без `API_ID` / `API_HASH`)
- 🔗 Поддержка Spotify ссылок: `track`, `album`, `playlist`
- 📦 Метаданные через Spotify Web API (Client Credentials)
- ⬇️ Скачивание аудио через `yt-dlp` + `ffmpeg`
- 🧰 Команды: `/start`, `/help`, `/thumb`, `/preview`

## 📋 Требования
- `Go 1.21+`
- `yt-dlp` в `PATH`
- `ffmpeg` в `PATH`
- Spotify API credentials

## 🛠 Установка зависимостей
```bash
sudo apt-get update -y && sudo apt-get install -y ffmpeg yt-dlp
```

Если нет `sudo`, поставь через свой пакетный менеджер или используй статические бинарники и добавь их в `PATH`.

## ⚙️ Переменные окружения
Пример лежит в файле `.env.example`.

```env
BOT_TOKEN=...                 # обязательно
SPOTIPY_CLIENT_ID=...         # обязательно (Spotify Client ID)
SPOTIPY_CLIENT_SECRET=...     # обязательно (Spotify Client Secret)
OWNER_ID=...                  # опционально
AUTH_CHATS="-100... 123"      # опционально, список chat_id через пробел
FIXIE_SOCKS_HOST=...          # опционально, socks5 host:port
AUDIO_FORMAT=mp3              # опционально: mp3/flac/m4a
AUDIO_QUALITY=320K            # опционально: 320K/192K/128K/0
YTDLP_BIN=...                 # опционально, путь к yt-dlp
```

Также поддерживаются `SPOTIFY_CLIENT_ID` / `SPOTIFY_CLIENT_SECRET`.

## 🚀 Запуск
```bash
cd /home/faline/spotify-downloader-go
cp .env.example .env
# заполнить .env
set -a && source .env && set +a
/usr/local/go/bin/go run ./cmd/bot
```

## 🧱 Сборка
```bash
cd /home/faline/spotify-downloader-go
/usr/local/go/bin/go build -o spotify-bot ./cmd/bot
```

## ▶️ Запуск бинарника
```bash
cd /home/faline/spotify-downloader-go
set -a && source .env && set +a
./spotify-bot
```

## 🧭 Команды BotFather
```text
start - Запуск бота
help - Помощь и список команд
thumb - Обложка Spotify (пример: /thumb <ссылка>)
preview - Превью трека Spotify (пример: /preview <ссылка>)
```

## 📝 Примечания
- Используется Telegram Bot API (HTTP), а не MTProto.
- `Redirect URI` в Spotify для этого проекта фактически не используется (достаточно валидного URL).
- Если видишь `Conflict: terminated by other getUpdates request`, значит запущено больше одного экземпляра бота с тем же токеном.
