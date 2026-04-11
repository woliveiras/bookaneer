# Build frontend first (Go embed needs web/dist)
FROM node:24-alpine AS frontend
WORKDIR /src/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

# Build backend (copy frontend dist for go:embed)
FROM golang:1.26.2-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/web/dist /src/web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(cat VERSION 2>/dev/null || echo dev) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o /app ./cmd/bookaneer

# Runtime — scratch for zero attack surface
FROM scratch

LABEL org.opencontainers.image.title="Bookaneer" \
      org.opencontainers.image.description="Self-hosted ebook collection manager" \
      org.opencontainers.image.url="https://github.com/woliveiras/bookaneer" \
      org.opencontainers.image.source="https://github.com/woliveiras/bookaneer"

COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /app /bookaneer

EXPOSE 9090
VOLUME ["/data", "/library"]

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD ["/bookaneer", "healthcheck"]

ENTRYPOINT ["/bookaneer"]
