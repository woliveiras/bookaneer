# Integrations

## Overview

Bookaneer integrates with external services for two functions: **searching** for books and **downloading** books. No external integration is required — the system works standalone as a library manager and web reader.

For digital-library lookup, Bookaneer supports legal open-access/public-domain sources such as Internet Archive, Gutendex (Project Gutenberg metadata API), and Open Library public scans.

```
                         ┌─────────────┐
                         │  BOOKANEER  │
                         └──────┬──────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                  │
        ┌─────▼──────┐  ┌──────▼──────┐   ┌──────▼──────┐
        │   SEARCH   │  │  DOWNLOAD   │   │  METADATA   │
        │            │  │             │   │             │
        └─────┬──────┘  └──────┬──────┘   └──────┬──────┘
              │                │                  │
      ┌───────┴────┐    ┌─────┴──────┐    ┌──────┴──────┐
      │ Prowlarr   │    │qBittorrent │    │ OpenLibrary │
      │ (optional) │    │Transmission│    │ GoogleBooks │
      │    OR      │    │  SABnzbd   │    │  HardCover  │
      │ Direct     │    │  NZBGet    │    └─────────────┘
      │ indexers   │    │ Blackhole  │
      └────────────┘    └────────────┘
```

---

## Indexers — Searching for books

### What are indexers?

Indexers are sites that index content available on Usenet or BitTorrent networks. Bookaneer queries these indexers to find ebooks you've marked as "Wanted".

Communication uses the standard **Newznab** (Usenet) and **Torznab** (Torrent) protocols — the same ones used by Readarr, Sonarr, Radarr and the entire *arr ecosystem.

### Option A: Direct indexer configuration

Each indexer is configured individually in Bookaneer with URL + API key:

```
Bookaneer
  ├─ Indexer 1: https://indexer1.com/api  (Torznab)
  ├─ Indexer 2: https://indexer2.com/api  (Newznab)
  └─ Indexer 3: https://indexer3.com/api  (Torznab)
```

**When to use:** if you have 1-2 indexers and don't use other *arr apps.

**Configuration in Bookaneer:**
```
Settings → Indexers → Add Indexer
  Type: Newznab or Torznab
  URL: https://your-indexer.com/api
  API Key: your-api-key
  Categories: Books (7020), Ebooks (7000)
```

### Option B: Via Prowlarr (recommended)

Prowlarr is a centralized indexer manager. You configure all your indexers **once in Prowlarr**, and it exposes them as Torznab/Newznab endpoints for any *arr app — including Bookaneer.

```
                    ┌──────────┐
                    │ Prowlarr │
                    │ :9696    │
                    └────┬─────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
    ┌─────▼─────┐  ┌────▼─────┐  ┌────▼─────┐
    │ Indexer 1 │  │ Indexer 2│  │ Indexer 3│
    │ (torrent) │  │ (usenet) │  │ (torrent)│
    └───────────┘  └──────────┘  └──────────┘

Bookaneer sees Prowlarr as a single "super indexer"
```

**When to use:** if you have multiple indexers or already use Prowlarr with other apps (Sonarr, Radarr, etc).

**Advantages of Prowlarr:**
- Configure indexers once, all apps benefit
- Prowlarr manages rate limits and retries per indexer
- Adding/removing indexers requires no changes in Bookaneer
- Centralized usage stats per indexer

**Setup:**

1. In Prowlarr, configure your indexers as usual

2. In Prowlarr, add Bookaneer as an app:
   ```
   Settings → Apps → Add → Generic (Newznab/Torznab)
   Or: Prowlarr syncs automatically if configured
   ```

3. In Bookaneer, add Prowlarr as an indexer:
   ```
   Settings → Indexers → Add Indexer
     Type: Torznab (or Newznab)
     URL: http://prowlarr:9696/{indexerId}/api
     API Key: (Prowlarr's api key)
     Categories: Books
   ```
   
   Or add each Prowlarr indexer individually — Prowlarr exposes each one as a separate endpoint.

### Search protocol (Newznab/Torznab)

When Bookaneer searches for a book, the request is:

```http
GET http://indexer/api?t=book&q=tolkien+lord+of+the+rings&apikey=xxx
```

The response is standard RSS/Newznab XML:

```xml
<rss>
  <channel>
    <item>
      <title>J.R.R. Tolkien - The Lord of the Rings (epub)</title>
      <size>2457600</size>
      <link>https://indexer/download/abc123</link>
      <attr name="seeders" value="42"/>
      <attr name="category" value="7020"/>
    </item>
  </channel>
</rss>
```

Bookaneer parses these results, applies the decision engine (quality matching, size, seeders) and selects the best candidate.

---

## Download Clients — Downloading books

### What are download clients?

After Bookaneer selects a result, it needs to send the download to a client that will actually download the file. Two types:

| Type | Base protocol | Supported clients |
|---|---|---|
| **Torrent** | BitTorrent | qBittorrent, Transmission, Deluge |
| **Usenet** | NNTP | SABnzbd, NZBGet |
| **Blackhole** | Filesystem | Any client that monitors a folder |

### qBittorrent

The most popular torrent client in the *arr ecosystem.

**Setup:**
```
Settings → Download Clients → Add → qBittorrent
  Host: qbittorrent (or IP)
  Port: 8080
  Username: admin
  Password: your-password
  Category: books
```

**Technical flow:**

```
1. SEND TORRENT
   POST http://qbittorrent:8080/api/v2/torrents/add
   Content-Type: multipart/form-data
   Body:
     urls=magnet:?xt=urn:btih:abc123...
     category=books
     savepath=/downloads/books

2. MONITOR PROGRESS (poll every ~60s)
   GET http://qbittorrent:8080/api/v2/torrents/info?category=books
   Response: [
     {
       "hash": "abc123",
       "name": "Tolkien - LOTR.epub",
       "progress": 1.0,        ← complete when = 1.0
       "state": "uploading",   ← seeding after completion
       "save_path": "/downloads/books"
     }
   ]

3. AFTER POST-PROCESSING
   # If configured to remove after import:
   POST http://qbittorrent:8080/api/v2/torrents/delete
   Body: hashes=abc123&deleteFiles=true
   
   # Or if configured to keep seeding: do nothing
```

**API auth:** qBittorrent uses cookie-based auth. Bookaneer logs in once and reuses the cookie:
```
POST http://qbittorrent:8080/api/v2/auth/login
Body: username=admin&password=secret
Response: Set-Cookie: SID=abc123
```

### Transmission

**Setup:**
```
Settings → Download Clients → Add → Transmission
  Host: transmission
  Port: 9091
  Username: admin
  Password: your-password
```

**Protocol:** JSON-RPC via HTTP.

```
POST http://transmission:9091/transmission/rpc
Headers:
  X-Transmission-Session-Id: <session-id>
Content-Type: application/json

# Add torrent:
{
  "method": "torrent-add",
  "arguments": {
    "filename": "magnet:?xt=urn:btih:abc123...",
    "download-dir": "/downloads/books"
  }
}

# List torrents:
{
  "method": "torrent-get",
  "arguments": {
    "fields": ["id", "name", "percentDone", "downloadDir", "status"]
  }
}
```

### SABnzbd (Usenet)

**Setup:**
```
Settings → Download Clients → Add → SABnzbd
  Host: sabnzbd
  Port: 8080
  API Key: your-sabnzbd-api-key
  Category: books
```

**Flow:**

```
# Send NZB:
POST http://sabnzbd:8080/api
  ?mode=addurl
  &name=https://indexer.com/download/abc123.nzb
  &cat=books
  &apikey=xxx

# Monitor:
GET http://sabnzbd:8080/api?mode=queue&output=json&apikey=xxx
Response: {
  "queue": {
    "slots": [{
      "nzo_id": "abc123",
      "filename": "Tolkien LOTR",
      "percentage": "100",
      "status": "Completed"
    }]
  }
}

# History (completed downloads):
GET http://sabnzbd:8080/api?mode=history&output=json&apikey=xxx
```

### NZBGet (Usenet)

**Setup:**
```
Settings → Download Clients → Add → NZBGet
  Host: nzbget
  Port: 6789
  Username: admin
  Password: your-password
  Category: books
```

**Protocol:** JSON-RPC.

```
POST http://nzbget:6789/jsonrpc
Authorization: Basic base64(admin:password)

# Send NZB by URL:
{
  "method": "append",
  "params": [
    "",                          # NZBFilename
    "https://indexer/nzb/123",   # URL
    "books",                     # Category
    0,                           # Priority
    false, false, "", 0, "SCORE" # Other params
  ]
}

# List downloads:
{
  "method": "listgroups",
  "params": []
}
```

### Embedded Downloader (no setup required)

When no download client is configured, Bookaneer falls back to the **Embedded Downloader** — a built-in HTTP client that downloads files directly from digital library sources (Anna's Archive, LibGen, Internet Archive) and saves them to your root library folder.

No configuration is needed. The download appears in the activity queue as `"Embedded Downloader"`.

**Limitation:** Many digital library sources protect download pages with Cloudflare or DDoS-Guard challenges. A plain HTTP request will receive a challenge page instead of the file, and the download will fail with a message like:

```
Cloudflare challenge detected — set flareSolverrUrl in config.yaml or BOOKANEER_FLARESOLVERR_URL
```

See [FlareSolverr bypass](#flaresolverr-optional-bypass-for-digital-libraries) below to resolve this.

---

### FlareSolverr — optional bypass for digital libraries

[FlareSolverr](https://github.com/FlareSolverr/FlareSolverr) is a headless-browser sidecar that solves Cloudflare Turnstile and DDoS-Guard challenges so Bookaneer can download files from protected sources.

**How it works:**

```
Embedded Downloader
  │
  ├─ 1. Plain HTTP GET → receives HTML challenge page
  │
  ├─ 2. Challenge detected (Cloudflare / DDoS-Guard)
  │
  ├─ 3. POST to FlareSolverr: { "cmd": "request.get", "url": "..." }
  │       FlareSolverr starts headless Chromium, solves challenge
  │       Returns: cookies (cf_clearance, etc) + user-agent
  │
  └─ 4. Retry HTTP GET with cookies → receives actual file ✓
```

**Setup:**

1. Add FlareSolverr to your Docker Compose:

```yaml
flaresolverr:
  image: ghcr.io/flaresolverr/flaresolverr:latest
  container_name: bookaneer-flaresolverr
  environment:
    - LOG_LEVEL=info
  restart: unless-stopped
  # No port mapping needed unless you want external access
```

2. Add `flareSolverrUrl` to `config.yaml`:

```yaml
flareSolverrUrl: http://flaresolverr:8191
```

Or set the environment variable:

```bash
BOOKANEER_FLARESOLVERR_URL=http://flaresolverr:8191
```

When configured, FlareSolverr is used for:
- The **Embedded Downloader** (automatic, on challenge detection)
- The **Anna's Archive** digital library provider (search + download page resolution)
- The **LibGen** digital library provider

FlareSolverr is **not required** — without it, downloads from protected sources will fail with a descriptive error in the activity queue. Public-domain sources (Internet Archive, Project Gutenberg, Wikisource) work without bypass.

**Resource requirements:** FlareSolverr runs a headless Chromium instance (~300–500 MB RAM). It is not suitable for devices with less than 1 GB RAM. On a Raspberry Pi 4 (4 GB) it works fine.

---

### Blackhole (any client)

For clients not directly supported. Bookaneer saves the .torrent or .nzb file to a folder monitored by the client.

**Setup:**
```
Settings → Download Clients → Add → Blackhole
  Torrent Folder: /blackhole/torrents
  NZB Folder: /blackhole/nzb
```

**Flow:**
```
1. Bookaneer saves file to /blackhole/torrents/book.torrent
2. Your torrent client monitors that folder and picks up the file
3. Client downloads to /downloads/
4. Bookaneer monitors /downloads/ for post-processing
```

---

## Complete flow: From "Want" to "Have"

```
┌── 1. User marks book as "Wanted" (UI or API)
│
├── 2. Bookaneer registers command: BookSearch { bookId: 42 }
│
├── 3. Scheduler executes (goroutine):
│     │
│     ├── 3a. Builds query: "tolkien lord of the rings epub"
│     │
│     ├── 3b. Searches each configured indexer (Prowlarr or direct):
│     │       GET prowlarr:9696/1/api?t=book&q=tolkien+lord+rings
│     │       GET prowlarr:9696/2/api?t=book&q=tolkien+lord+rings
│     │       → Receives N XML results
│     │
│     ├── 3c. Decision Engine filters and ranks:
│     │       ✓ Name matches book? (fuzzy match ≥ 80%)
│     │       ✓ Accepted format? (epub > pdf > mobi, per quality profile)
│     │       ✓ Reasonable size? (not 50 GB)
│     │       ✓ Enough seeders? (> configured minimum)
│     │       → Selects best result
│     │
│     ├── 3d. Sends to download client:
│     │       POST qbittorrent:8080/api/v2/torrents/add { magnet }
│     │       → Book status: "Wanted" → "Snatched"
│     │       → on_grab notification fires (webhook, discord, email)
│     │       → WebSocket → frontend updates
│     │
│     └── 3e. Records in history: { event: "grabbed", source: "indexer1" }
│
├── 4. Download monitor (periodic goroutine, ~60s):
│     │
│     ├── Poll qBittorrent:
│     │   GET /api/v2/torrents/info?category=books
│     │
│     ├── Completed download detected
│     │
│     └── 4a. Post-processing:
│           ├── Identifies file: /downloads/books/Tolkien-LOTR.epub
│           ├── Matches with book in database (fuzzy title match)
│           ├── Rename: → /library/J.R.R. Tolkien/The Lord of the Rings.epub
│           ├── Extracts/downloads cover art
│           ├── Writes metadata.opf
│           ├── Book status: "Snatched" → "Have"
│           ├── Records in history: { event: "downloaded" }
│           ├── on_download notification fires
│           └── WebSocket → UI updates
│
└── 5. Book available:
      ├── In the library (filesystem)
      ├── In the web reader (/reader/:id)
      └── In the OPDS feed (reading apps)
```

---

## Shared volume: /downloads

The integration point between download clients and Bookaneer is a **shared filesystem volume**.

```
download client saves here
        │
        ▼
  /downloads/books/
        │
        ▼
Bookaneer reads from here (post-processing)
        │
        ▼
  /library/Author Name/Book Title.epub
```

Both containers need access to the same path. In Docker:

```yaml
services:
  bookaneer:
    volumes:
      - /media/downloads:/downloads  # read
      - /media/books:/library        # write

  qbittorrent:
    volumes:
      - /media/downloads:/downloads  # write
```

If internal paths differ between containers, use **Remote Path Mappings** in Bookaneer:

```
Settings → Download Clients → Remote Path Mappings
  Host: qbittorrent
  Remote Path: /data/downloads/    (path inside qBittorrent container)
  Local Path: /downloads/          (path inside Bookaneer container)
```

---

## Reference Docker Compose

```yaml
# docker-compose.yml — Full setup with Prowlarr + qBittorrent
# Bookaneer works standalone without the other services

services:
  # ─── BOOKANEER ─────────────────────────────────────────
  # Required. Works standalone as library manager + reader.
  bookaneer:
    image: bookaneer
    container_name: bookaneer
    ports:
      - "9090:9090"
    volumes:
      - ./bookaneer-data:/data        # SQLite, config, cache
      - /media/books:/library         # Ebook library
      - /media/downloads:/downloads   # Downloads (shared)
    restart: unless-stopped

  # ─── PROWLARR (optional) ───────────────────────────────
  # Centralized indexer manager.
  # Not needed if indexers are configured directly.
  prowlarr:
    image: lscr.io/linuxserver/prowlarr:latest
    container_name: prowlarr
    ports:
      - "9696:9696"
    volumes:
      - ./prowlarr-config:/config
    restart: unless-stopped

  # ─── qBITTORRENT (optional) ────────────────────────────
  # Download client for torrents.
  # Not needed if using only blackhole or Usenet.
  qbittorrent:
    image: lscr.io/linuxserver/qbittorrent:latest
    container_name: qbittorrent
    ports:
      - "8080:8080"
    volumes:
      - ./qbittorrent-config:/config
      - /media/downloads:/downloads   # Shared with Bookaneer
    restart: unless-stopped

  # ─── SABnzbd (optional) ────────────────────────────────
  # Download client for Usenet.
  # Not needed if using only torrents.
  # sabnzbd:
  #   image: lscr.io/linuxserver/sabnzbd:latest
  #   container_name: sabnzbd
  #   ports:
  #     - "8081:8080"
  #   volumes:
  #     - ./sabnzbd-config:/config
  #     - /media/downloads:/downloads
  #   restart: unless-stopped
```

---

## Summary of APIs consumed by Bookaneer

| Service | Protocol | Default port | Required? |
|---|---|---|---|
| **OpenLibrary** | REST JSON | 443 (HTTPS) | Yes (primary metadata) |
| **GoogleBooks** | REST JSON | 443 (HTTPS) | No (metadata fallback) |
| **HardCover** | REST JSON | 443 (HTTPS) | No (metadata fallback) |
| **Prowlarr** | Newznab/Torznab XML | 9696 | **No** (optional proxy) |
| **Direct indexer** | Newznab/Torznab XML | varies | For automatic search |
| **qBittorrent** | REST JSON | 8080 | For torrent downloads |
| **Transmission** | JSON-RPC | 9091 | Alternative to qBittorrent |
| **Deluge** | JSON-RPC | 8112 | Alternative to qBittorrent |
| **SABnzbd** | REST query params | 8080 | For Usenet downloads |
| **NZBGet** | JSON-RPC | 6789 | Alternative to SABnzbd |
