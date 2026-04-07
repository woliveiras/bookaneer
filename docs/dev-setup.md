# Dev Setup Guide — Bookaneer

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| **mise** | latest | Version manager for Go, Node, pnpm |
| **goose** | CLI | Database migrations |

### Install mise

```bash
# macOS
brew install mise

# Linux (script)
curl https://mise.run | sh

# Or see: https://mise.jdx.dev/getting-started.html
```

Add to your shell (bash/zsh):
```bash
echo 'eval "$(mise activate)"' >> ~/.zshrc
cd bookaneer
mise trust
source ~/.zshrc
```

### Install project tools

```bash
mise install          # Installs Go 1.26, Node 24, pnpm 10
go install github.com/pressly/goose/v3/cmd/goose@latest
```

> **Note:** The project includes a `mise.toml` file that pins exact versions. All contributors will use the same toolchain.

---

## Project structure

```
bookaneer/
├── cmd/bookaneer/         # Entrypoint
│   └── main.go
├── internal/
│   ├── api/v1/            # Echo route handlers
│   ├── config/            # YAML parsing + defaults
│   ├── core/              # Domain models
│   ├── database/          # SQLite connection, migrations
│   ├── download/          # Download client integrations
│   ├── metadata/          # OpenLibrary, Google Books
│   ├── notification/      # Webhook, Discord, email, etc.
│   ├── opds/              # OPDS catalog server
│   ├── reader/            # EPUB parser, web reader API
│   ├── scheduler/         # Job queue + scheduler
│   └── search/            # Indexer queries (Newznab/Torznab)
├── migrations/            # SQL migrations (goose)
├── web/                   # React frontend (Vite)
├── docs/                  # Documentation
├── data/                  # Runtime: DB, logs, backups (gitignored)
└── Makefile
```

---

## Initial setup

### 1. Clone

```bash
git clone git@github.com:woliveiras/bookaneer.git
cd bookaneer
```

### 2. Backend

```bash
# Download Go dependencies
go mod download

# Run migrations
mkdir -p data
goose -dir migrations sqlite3 ./data/bookaneer.db up

# Run backend in dev mode (hot reload with air, or directly)
go run cmd/bookaneer/main.go

# Or with air (recommended for dev):
go install github.com/air-verse/air@latest
air
```

The backend starts at `http://localhost:8787`.

### 3. Frontend

```bash
cd web
pnpm install
pnpm dev
```

The Vite dev server starts at `http://localhost:5173` with proxy to the backend.

### 4. Both (recommended)

Use the Makefile:

```bash
make dev    # Starts backend (air) + frontend (vite) in parallel
```

---

## Environment variables

| Variable | Default | Description |
|----------|---------|-----------|
| `BOOKANEER_PORT` | `8787` | HTTP port |
| `BOOKANEER_DATA_DIR` | `./data` | Data directory (DB, logs, backups) |
| `BOOKANEER_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

The main config file lives at `$BOOKANEER_DATA_DIR/config.yaml`. If it doesn't exist, it is created with defaults on first boot.

---

## Useful commands

```bash
# Build
make build                 # Compila binary em ./dist/bookaneer
make build-all             # Cross-compile linux/amd64, linux/arm64, darwin/arm64

# Test
make test                  # go test ./...
make test-race             # go test -race ./...
make test-cover            # With coverage report

# Lint
make lint                  # golangci-lint run

# Database
make migrate-up            # goose up
make migrate-down          # goose down (rollback 1)
make migrate-create NAME=add_something  # New migration file

# Frontend
make web-build             # pnpm --prefix web build
make web-lint              # pnpm --prefix web lint

# Docker
make docker                # docker build -t bookaneer:dev .
make docker-run            # docker run com volumes mapeados
```

---

## Makefile

```makefile
.PHONY: build build-all dev test lint migrate-up migrate-down docker

BINARY    := bookaneer
DATA_DIR  := ./data
DB_PATH   := $(DATA_DIR)/bookaneer.db
GOOSE     := goose -dir migrations sqlite3 $(DB_PATH)

# --- Build ---
build:
	go build -o dist/$(BINARY) ./cmd/bookaneer

build-all:
	GOOS=linux  GOARCH=amd64 go build -o dist/$(BINARY)-linux-amd64   ./cmd/bookaneer
	GOOS=linux  GOARCH=arm64 go build -o dist/$(BINARY)-linux-arm64   ./cmd/bookaneer
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY)-darwin-arm64  ./cmd/bookaneer

# --- Dev ---
dev:
	@mkdir -p $(DATA_DIR)
	@$(GOOSE) up
	@trap 'kill 0' EXIT; \
		air & \
		(cd web && pnpm dev) & \
		wait

# --- Test ---
test:
	go test ./...

test-race:
	go test -race ./...

test-cover:
	go test -coverprofile=$(DATA_DIR)/coverage.out ./...
	go tool cover -html=$(DATA_DIR)/coverage.out -o $(DATA_DIR)/coverage.html
	@echo "Coverage report: $(DATA_DIR)/coverage.html"

# --- Lint ---
lint:
	golangci-lint run ./...

# --- Database ---
migrate-up:
	@mkdir -p $(DATA_DIR)
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-create:
	$(GOOSE) create $(NAME) sql

# --- Frontend ---
web-build:
	pnpm --prefix web build

web-lint:
	pnpm --prefix web lint

# --- Docker ---
docker:
	docker build -t $(BINARY):dev .

docker-run:
	docker run --rm -it \
		-p 8787:8787 \
		-v $(PWD)/data:/data \
		-v $(PWD)/library:/library \
		$(BINARY):dev
```

---

## Air config (.air.toml)

```toml
root = "."
tmp_dir = "data/tmp"

[build]
  bin = "./data/tmp/bookaneer"
  cmd = "go build -o ./data/tmp/bookaneer ./cmd/bookaneer"
  delay = 1000
  exclude_dir = ["data", "web", "node_modules", "dist", ".git"]
  exclude_regex = ["_test\\.go$"]
  include_ext = ["go", "yaml", "sql"]
  kill_delay = 500

[log]
  time = false

[misc]
  clean_on_exit = true
```

---

## Vite proxy config (web/vite.config.ts)

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8787',
      '/opds': 'http://localhost:8787',
      '/ws': {
        target: 'ws://localhost:8787',
        ws: true,
      },
    },
  },
})
```

---

## Docker (production)

```dockerfile
# --- Build stage ---
FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bookaneer ./cmd/bookaneer

FROM node:24-alpine AS frontend
WORKDIR /src/web
COPY web/package.json web/pnpm-lock.yaml* ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

# --- Runtime stage ---
FROM scratch
COPY --from=backend /bookaneer /bookaneer
COPY --from=frontend /src/web/dist /web/dist
EXPOSE 8787
VOLUME ["/data", "/library"]
ENTRYPOINT ["/bookaneer"]
```

> **Note:** In production, the frontend is embedded in the Go binary via `embed.FS`.
> The multi-stage Dockerfile above is the alternative for when you don't want to recompile Go on every frontend change.
> The final build uses `go:embed web/dist` and produces a single binary.

---

## First boot

On the first run:

1. Creates `data/bookaneer.db` and runs migrations
2. Generates a random API key
3. Creates admin user (username/password set via flag or interactive prompt)
4. Open `http://localhost:8787` — setup wizard:
   - Configure root folder
   - (Optional) Configure download client
   - (Optional) Configure indexer
   - Add first author/book
