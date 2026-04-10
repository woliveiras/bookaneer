# Bookaneer

A self-hosted ebook collection manager. Combines features from Readarr and LazyLibrarian into a single binary (~15 MB Docker image) targeting home users, NAS devices, and Raspberry Pi.

<p align="center">
  <img src="assets/bookaneer_200.png" alt="Bookaneer Mascot" width="200">
</p>

## Features

- **Library management** — Scan existing ebooks, organize by author/series, fetch metadata
- **Web reader** — Read EPUBs directly in your browser (powered by Foliate-js)
- **OPDS catalog** — Access your library from any OPDS-compatible reading app
- **Automatic search** — Find books via Newznab/Torznab indexers (optional)
- **Download integration** — qBittorrent, Transmission, SABnzbd, NZBGet (optional)
- **Notifications** — Webhook, Discord, email, and more

## Quick Start

```bash
docker run -d \
  --name bookaneer \
  -p 9090:9090 \
  -v ./data:/data \
  -v /path/to/books:/library \
  bookaneer/bookaneer:latest
```

Open `http://localhost:9090` to access the web UI.

## Documentation

- [Developer Setup](docs/dev-setup.md)
- [Architecture Overview](docs/architecture-overview.md)
- [API Specification](docs/api-spec.md)
- [Integrations](docs/integrations.md)

## Tech Stack

- **Backend**: Go 1.26+, Echo v4, SQLite (WAL mode)
- **Frontend**: React 19, Vite, TanStack Query, shadcn/ui
- **Web Reader**: Foliate-js (MIT)

## Legal Disclaimer

**Bookaneer is intended for managing legally obtained ebooks.** The developers do not condone, encourage, or promote copyright infringement or piracy in any form.

This software:
- Is designed to organize and read ebooks you already own
- Integrates with standard protocols (Newznab/Torznab, OPDS) used by many legitimate services
- Does not host, distribute, or provide access to copyrighted content
- **Does not circumvent or remove DRM** — files with DRM protection will not open

**Users are solely responsible for ensuring their use of this software complies with all applicable copyright laws in their jurisdiction.**

The authors and contributors of this project assume no liability for any misuse of this software.

## License

This project is licensed under the GNU General Public License v3.0 — see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please read [AGENTS.md](AGENTS.md) for project guidelines and conventions before submitting PRs.
