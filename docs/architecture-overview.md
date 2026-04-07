# Architecture — Overview

## How Bookaneer fits in the ecosystem

```
┌─────────────────────────────────────────────────────────────┐
│                        BOOKANEER                            │
│                                                             │
│  ┌──────────┐   ┌──────────┐   ┌───────────────────────┐  │
│  │ Metadata  │   │ Search   │   │  Download Manager     │  │
│  │ Providers │   │ Engine   │   │  (queue + post-proc)  │  │
│  └─────┬────┘   └─────┬────┘   └──────────┬────────────┘  │
│        │               │                   │                │
│  ┌─────┴────┐   ┌─────┴────┐   ┌──────────┴────────────┐  │
│  │ Web      │   │ OPDS     │   │  Web Reader            │  │
│  │ UI       │   │ Server   │   │  (Foliate-js)          │  │
│  └──────────┘   └──────────┘   └────────────────────────┘  │
└────────┬────────────────┬──────────────────┬────────────────┘
         │                │                  │
    ┌────▼────┐    ┌──────▼──────┐    ┌──────▼──────┐
    │Metadata │    │  Indexers   │    │  Download   │
    │ APIs    │    │  (search)   │    │  Clients    │
    └─────────┘    └─────────────┘    └─────────────┘
```

### External components

| Component | Required? | What it does |
|---|---|---|
| **Metadata APIs** (OpenLibrary, GoogleBooks, HardCover) | Yes (at least 1) | Provides author and book info |
| **Indexers** (via Prowlarr or direct) | For automatic search | Searches for book torrents/NZBs |
| **Download clients** (qBittorrent, SABnzbd, etc) | For automatic download | Downloads the found files |
| **Prowlarr** | **No** | Centralized proxy for multiple indexers |

Bookaneer works standalone as a **library manager + web reader**. Indexers and download clients are only needed for automatic search and download.

### Operation modes

**Mode 1 — Library + Reader only (zero external setup):**
- Scan existing ebooks on disk
- Metadata enrichment via OpenLibrary/GoogleBooks
- Web reader for in-browser reading
- OPDS feed for reading apps

**Mode 2 — With automatic download:**
- Everything from Mode 1, plus:
- Configure indexers (direct or via Prowlarr) to search for books
- Configure download client to download
- Automatic post-processing (rename, organize, metadata)

See [integrations.md](integrations.md) for details on each integration.
