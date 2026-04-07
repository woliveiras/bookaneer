.PHONY: build build-all dev test test-race test-cover lint migrate-up migrate-down docker docker-run clean web-build web-lint help

BINARY    := bookaneer
VERSION   := $(shell cat VERSION 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"
DATA_DIR  := ./data
DB_PATH   := $(DATA_DIR)/bookaneer.db
GOOSE     := goose -dir migrations sqlite3 $(DB_PATH)

# --- Build ---
build:
	@echo "Building $(BINARY)..."
	@mkdir -p dist
	go build $(LDFLAGS) -o dist/$(BINARY) ./cmd/bookaneer

build-all:
	@echo "Cross-compiling..."
	@mkdir -p dist
	GOOS=linux  GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64   ./cmd/bookaneer
	GOOS=linux  GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64   ./cmd/bookaneer
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64  ./cmd/bookaneer

# --- Dev ---
dev:
	@mkdir -p $(DATA_DIR)
	@if command -v air > /dev/null 2>&1; then \
		trap 'kill 0' EXIT; \
		air & \
		(cd web && pnpm dev) & \
		wait; \
	else \
		echo "air not installed. Run: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

dev-backend:
	@mkdir -p $(DATA_DIR)
	go run ./cmd/bookaneer

dev-frontend:
	cd web && pnpm dev

# --- Test ---
test:
	go test ./...

test-race:
	go test -race ./...

test-cover:
	@mkdir -p $(DATA_DIR)
	go test -coverprofile=$(DATA_DIR)/coverage.out ./...
	go tool cover -html=$(DATA_DIR)/coverage.out -o $(DATA_DIR)/coverage.html
	@echo "Coverage report: $(DATA_DIR)/coverage.html"

# --- Lint ---
lint:
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed"; \
		exit 1; \
	fi
	cd web && pnpm lint

# --- Database ---
migrate-up:
	@mkdir -p $(DATA_DIR)
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=add_something"; \
		exit 1; \
	fi
	$(GOOSE) create $(NAME) sql

# --- Frontend ---
web-build:
	cd web && pnpm build

web-lint:
	cd web && pnpm lint

web-install:
	cd web && pnpm install

# --- Docker ---
docker:
	docker build -t $(BINARY):dev .

docker-run:
	docker run --rm -it \
		-p 9090:9090 \
		-v $(PWD)/data:/data \
		-v $(PWD)/library:/library \
		$(BINARY):dev

# --- Clean ---
clean:
	rm -rf dist/
	rm -rf web/dist/
	rm -rf $(DATA_DIR)/tmp/

# --- Help ---
help:
	@echo "Bookaneer Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make dev            Start backend (air) + frontend (vite) with hot reload"
	@echo "  make dev-backend    Start only backend"
	@echo "  make dev-frontend   Start only frontend"
	@echo "  make build          Build binary to dist/"
	@echo "  make build-all      Cross-compile for linux/darwin"
	@echo "  make test           Run tests"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make test-cover     Run tests with coverage report"
	@echo "  make lint           Run golangci-lint and eslint"
	@echo "  make migrate-up     Run database migrations"
	@echo "  make migrate-down   Rollback last migration"
	@echo "  make migrate-create NAME=xxx  Create new migration"
	@echo "  make web-build      Build frontend for production"
	@echo "  make docker         Build Docker image"
	@echo "  make docker-run     Run Docker container"
	@echo "  make clean          Remove build artifacts"
