# Build backend
FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(cat VERSION 2>/dev/null || echo dev) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o /app ./cmd/bookaneer

# Build frontend
FROM node:24-alpine AS frontend
WORKDIR /src/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
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
