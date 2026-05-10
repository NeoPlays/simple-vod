# Reel

A lightweight self-hosted Video-on-Demand service. Drop `.mp4` files in a folder, sync them, and stream from a clean dark UI — no transcoding, no bloat.

**Stack:** Go 1.25 · SQLite (pure Go, no CGO) · Alpine.js · Pico.css

---

## Quick start (Docker)

```bash
cp .env.example .env
# Edit .env — at minimum change PASSWORD_PEPPER
docker compose up -d
```

Open [http://localhost:8080](http://localhost:8080). The first user to register becomes admin.

Place `.mp4` files in the `videos` Docker volume (or bind-mount a local folder) and hit **Sync** in the UI to import them.

### Bind-mounting a local video directory

```yaml
# docker-compose.yml override
services:
  app:
    volumes:
      - db:/data/db
      - /path/to/your/videos:/data/videos
```

---

## Development

### Prerequisites

- Go 1.25+

### Run

```bash
cd backend
go run ./cmd/server
```

The server starts on `:8080` and serves the frontend from `../frontend`.

### Build

```bash
cd backend
go build ./...
```

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `./db/sqlite.db` | SQLite database file path |
| `VIDEO_DIR` | `./data/videos` | Directory scanned for `.mp4` files |
| `FRONTEND_DIR` | `../frontend` | Directory served as static files |
| `PASSWORD_PEPPER` | `default-pepper` | Prepended to passwords before bcrypt — **change in production** |

For local dev, export variables directly or use [direnv](https://direnv.net/):

```bash
export PASSWORD_PEPPER=my-local-pepper
go run ./cmd/server
```

---

## Docker Hub

Images are published automatically on every push to `main`.

```bash
docker pull neoplays/simple-vod:latest
```

### Required repository secrets

| Secret | Description |
|---|---|
| `DOCKERHUB_USERNAME` | Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token (read/write) |

Add them under **Settings → Secrets and variables → Actions** in your GitHub repository.

---

## Auth

- First registered user is automatically promoted to admin — no token needed.
- Subsequent registrations require an invite token created by an admin at `/tokens.html`.
- Tokens expire after 7 days.

## License

[MIT](LICENSE)
