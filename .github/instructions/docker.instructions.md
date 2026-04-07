---
description: "Use when writing, reviewing, or modifying Dockerfile, docker-compose files, or container configuration. Covers Docker security, multi-stage builds, scratch images, minimal attack surface, and container best practices for Go applications."
applyTo: "Dockerfile, docker-compose*.yml, .dockerignore"
---

# Docker & Container Infrastructure

## Multi-Stage Build Pattern

Always use multi-stage builds to produce minimal images:

```dockerfile
# Build backend
FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app ./cmd/bookaneer

# Build frontend
FROM node:24-alpine AS frontend
WORKDIR /src/web
COPY web/package.json web/pnpm-lock.yaml* ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

# Runtime — scratch for zero attack surface
FROM scratch
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /app /bookaneer
COPY --from=frontend /src/web/dist /web/dist
EXPOSE 9090
VOLUME ["/data", "/library"]
ENTRYPOINT ["/bookaneer"]
```

## Image Security

- **Base image**: `scratch` for production — zero shell, zero package manager, zero attack surface
- **CGO_ENABLED=0**: Pure Go binary, no C dependencies, runs on scratch
- **CA certificates**: Copy from build stage for HTTPS calls to external APIs
- **No root**: The scratch image has no users. If using `alpine` base, add: `RUN adduser -D -u 1000 app && USER app`
- **No secrets in images**: Never `COPY` `.env`, config files with credentials, or `data/` directory
- **Read-only filesystem**: Run with `--read-only` and mount only `/data` and `/library` as writable
- **Pin versions**: Use specific tags (`golang:1.26-alpine`), not `latest`
- **Strip binary**: `-ldflags="-s -w"` removes debug symbols (~30% smaller)
- **Scan images**: Use `docker scout` or `trivy` in CI

## .dockerignore

Always include a `.dockerignore` to speed up builds and avoid leaking data:

```
.git
data/
node_modules/
dist/
*.db
*.db-wal
*.db-shm
.env
.air.toml
```

## Docker Compose (Development)

```yaml
services:
  bookaneer:
    build: .
    ports:
      - "9090:9090"
    volumes:
      - ./data:/data        # DB, config, logs
      - ./library:/library   # Book files
    environment:
      - BOOKANEER_LOG_LEVEL=debug
    restart: unless-stopped
    # Security hardening
    read_only: true
    tmpfs:
      - /tmp
    security_opt:
      - no-new-privileges:true
```

## Volume Conventions

- `/data` — Application data: SQLite DB, config.yaml, logs, backups, cover cache
- `/library` — Book files: organized by author/title
- Never store data in the image — always use volumes
- Use named volumes in production, bind mounts in development

## Health Check

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD ["/bookaneer", "healthcheck"]
```

The binary should support a `healthcheck` subcommand that does an HTTP GET to `localhost:9090/api/v1/system/health` and exits 0/1.

## Resource Limits

Recommend in documentation:

```yaml
deploy:
  resources:
    limits:
      memory: 256M    # More than enough for ~25 MB idle
      cpus: "1.0"
```

## Logging

- Log to stdout/stderr — Docker captures it
- JSON format in production for `docker logs` parsing
- Never log to files inside the container (use volumes if file logging is needed)
- Include container-friendly log fields: no ANSI colors when `NO_COLOR` is set or stdout is not a TTY

## Build Args & Labels

```dockerfile
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT_SHA

LABEL org.opencontainers.image.title="Bookaneer"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.source="https://github.com/woliveiras/bookaneer"
LABEL org.opencontainers.image.created="${BUILD_TIME}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
```

## Cross-Platform Builds

Pure Go (`CGO_ENABLED=0`) supports trivial multi-arch:

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t bookaneer:latest .
```

No special toolchain needed — Go cross-compiles natively.

## Anti-patterns

- Using `alpine` or `ubuntu` as runtime when `scratch` works (Go static binary)
- Running as root inside the container
- `COPY . .` without `.dockerignore` (leaks `.git`, `data/`, env files)
- Installing tools in the runtime image (curl, wget, bash "for debugging")
- Using `latest` tag for base images
- Storing DB or config inside the image layer
- `docker run --privileged` or `--cap-add=ALL`
- Exposing unnecessary ports
