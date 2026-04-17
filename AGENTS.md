# Bookaneer — Project Guidelines

## Overview

Bookaneer is a self-hosted ebook collection manager written in Go. It combines features from Readarr and LazyLibrarian into a single binary (~15 MB Docker image) targeting home users, NAS devices, and Raspberry Pi.

## Stack

- **Backend**: Go 1.26+, Echo v4, modernc.org/sqlite (pure Go, zero CGo)
- **Frontend**: React 19, Vite, TanStack Query, TanStack Router, shadcn/ui, Tailwind CSS
- **Web Reader**: Foliate-js (EPUB renderer, MIT)
- **Database**: SQLite WAL mode only
- **Jobs**: SQLite-backed command queue + goroutine dispatcher
- **Build**: Single binary via `embed.FS` (Go embeds frontend)
- **Container**: Docker scratch base image

## Code Style & Architecture

Detailed rules live in the instruction files — see links below. Key references:

- **Go**: [.github/instructions/golang.instructions.md](.github/instructions/golang.instructions.md) — Effective Go + Uber Go Style Guide
- **Frontend**: [.github/instructions/react-frontend.instructions.md](.github/instructions/react-frontend.instructions.md) — TypeScript strict, TanStack, shadcn/ui, a11y, state management
- **XState**: [.github/instructions/xstate.instructions.md](.github/instructions/xstate.instructions.md) — XState v5 machines, `createMachine` patterns, testing with `createActor`
- **Zustand**: [.github/instructions/zustand.instructions.md](.github/instructions/zustand.instructions.md) — Zustand v5 stores, middleware, immer, persistence
- **Zod**: [.github/instructions/zod.instructions.md](.github/instructions/zod.instructions.md) — Zod v4 schemas, type inference, runtime validation in `fetchAPI`
- **Architecture**: [.github/instructions/architecture.instructions.md](.github/instructions/architecture.instructions.md) — Layered architecture, dependency rules, service patterns
- **SQLite**: [.github/instructions/sqlite.instructions.md](.github/instructions/sqlite.instructions.md) — WAL mode, parameterized queries, migrations
- **Docker**: [.github/instructions/docker.instructions.md](.github/instructions/docker.instructions.md) — Multi-stage build, scratch base, security
- **Security**: [.github/instructions/security.instructions.md](.github/instructions/security.instructions.md) — OWASP Top 10:2025, input validation, auth

## Build and Test

```bash
make dev          # Backend (air) + Frontend (vite) with hot reload
make build        # Single binary → dist/bookaneer
make test         # go test ./... && pnpm --prefix web test
make test-race    # go test -race ./...
make lint         # golangci-lint + pnpm --prefix web lint
make docker       # Docker scratch image
```

## Documentation

- [docs/architecture-overview.md](docs/architecture-overview.md) — Architecture diagrams
- [docs/integrations.md](docs/integrations.md) — External service integrations
- [docs/api-spec.md](docs/api-spec.md) — REST API specification
- [docs/dev-setup.md](docs/dev-setup.md) — Developer setup guide
- [migrations/001_initial_schema.sql](migrations/001_initial_schema.sql) — Database schema
