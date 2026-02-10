# StreamBox

Self-hosted torrent streaming web app. Browse movies, find torrents, and stream directly in your browser — no waiting for full downloads.

## Features

- **Movie browsing** — Trending, popular, and search powered by TMDB (Russian metadata)
- **Torrent search** — Rutracker (Russian dubs) with YTS (English) fallback
- **Real-time streaming** — Stream while downloading, MKV/AVI auto-transcoded to MP4 via FFmpeg
- **Custom video player** — Seeking, playback speed (0.5x–2x), Picture-in-Picture, keyboard shortcuts
- **Audio track selection** — Switch between audio tracks in multi-audio MKV files (e.g. Russian dub / original)
- **Subtitles** — OpenSubtitles integration with Russian and English options
- **Watch history** — Progress auto-saved, continue watching from where you left off
- **Mobile-friendly** — Double-tap seek, responsive controls

## Tech Stack

| Layer | Stack |
|-------|-------|
| Backend | Go, Gin, anacrolix/torrent, SQLite (CGo-free) |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4 |
| Transcoding | FFmpeg (fragmented MP4 pipe) |
| Deployment | Docker, multi-stage build, ~50MB image |

## Quick Start

### Docker (recommended)

```bash
cp .env.example .env
# Edit .env with your API keys

docker compose up -d
```

Open `http://localhost:8080`

### Development

**Prerequisites:** Go 1.23+, Node 20+, FFmpeg

```bash
# Backend
cd backend
export $(grep -v '^#' ../.env | xargs)
go run ./cmd/server/

# Frontend (separate terminal)
cd frontend
npm install
npm run dev
```

Frontend dev server runs on `http://localhost:5173` and proxies `/api` to the backend.

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `TMDB_API_KEY` | Yes | [TMDB API key](https://www.themoviedb.org/settings/api) |
| `RUTRACKER_USERNAME` | Yes | Rutracker account username |
| `RUTRACKER_PASSWORD` | Yes | Rutracker account password |
| `RUTRACKER_MIRROR` | No | Mirror domain (default: `rutracker.org`) |
| `OPENSUBTITLES_API_KEY` | No | [OpenSubtitles API key](https://www.opensubtitles.com/consumers) |
| `PORT` | No | Server port (default: `8080`) |
| `DATA_DIR` | No | Database and cache directory (default: `./data`) |
| `MAX_CACHE_GB` | No | Torrent cache size limit (default: `50`) |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` / `K` | Play / Pause |
| `Left` / `Right` | Seek -10s / +10s |
| `Shift+Left` / `Shift+Right` | Seek -30s / +30s |
| `Up` / `Down` | Volume up / down |
| `F` | Toggle fullscreen |
| `M` | Toggle mute |
| `C` | Toggle subtitles |
| `<` / `>` | Decrease / increase playback speed |

## Architecture

```
browser → Vite (dev) or static files (prod)
  ↓
/api/* → Go backend (Gin)
  ├── /movies/*      → TMDB proxy
  ├── /torrents/*    → Rutracker / YTS search
  ├── /stream/*      → Torrent → FFmpeg → HTTP chunked
  ├── /subtitles/*   → OpenSubtitles proxy (SRT→WebVTT)
  └── /history/*     → SQLite watch history
```

For MKV files, the backend pipes torrent data through FFmpeg:
```
torrent reader → FFmpeg (-c:v copy -c:a aac -f mp4) → HTTP response
```

Video/audio are not re-encoded (video is copied, only audio is transcoded to AAC), so CPU usage stays low.

## License

MIT
