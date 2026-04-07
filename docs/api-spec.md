# API Spec — Bookaneer v1

Base URL: `/api/v1`

Auth: all routes require `X-Api-Key` header or `?apikey=` query param.
Authentication routes (`/auth/*`) are the exception.

Format: JSON. Dates in ISO 8601. IDs are integers auto-increment.

## Conventions

- `GET` list endpoints support: `?page=1&pageSize=20&sortKey=title&sortDir=asc`
- List responses return: `{ "page": 1, "pageSize": 20, "totalRecords": 150, "records": [...] }`
- Error responses: `{ "error": "message", "details": [...] }`
- `201 Created` returns the created object
- `204 No Content` for deletes and updates without body
- `400` for validation, `401` for auth, `404` for not found, `409` for conflict

---

## Auth

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/login` | Login with username/password, returns API key |
| `POST` | `/auth/logout` | Invalidate session |
| `GET` | `/auth/me` | Current user |

### POST /auth/login
```json
// Request
{ "username": "admin", "password": "secret" }

// Response 200
{ "username": "admin", "role": "admin", "apiKey": "abc123..." }
```

---

## Authors

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/author` | List authors |
| `GET` | `/author/:id` | Author detail (includes books) |
| `POST` | `/author` | Add author (by name or foreign ID) |
| `PUT` | `/author/:id` | Update author |
| `DELETE` | `/author/:id` | Remove author and their books |

### GET /author
Query params: `?monitored=true&status=active&sortKey=name&sortDir=asc`

```json
// Response 200
{
  "page": 1, "pageSize": 20, "totalRecords": 42,
  "records": [
    {
      "id": 1,
      "name": "J.R.R. Tolkien",
      "sortName": "Tolkien, J.R.R.",
      "foreignAuthorId": "OL26320A",
      "overview": "English writer and philologist...",
      "imageUrl": "/api/v1/media/author/1/cover",
      "status": "active",
      "monitored": true,
      "path": "/library/J.R.R. Tolkien",
      "bookCount": 12,
      "bookFileCount": 8,
      "sizeOnDisk": 52428800,
      "addedAt": "2026-04-06T10:00:00Z"
    }
  ]
}
```

### POST /author
```json
// Request — by search
{ "foreignAuthorId": "OL26320A", "monitored": true, "rootFolderPath": "/library" }

// Request — by name (fetches metadata automatically)
{ "name": "J.R.R. Tolkien", "monitored": true, "rootFolderPath": "/library" }

// Response 201 — returns full author with books populated via metadata
```

### PUT /author/:id
```json
// Request (partial fields accepted)
{ "monitored": false, "status": "paused" }
```

---

## Books

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/book` | List books |
| `GET` | `/book/:id` | Book detail (includes editions, files) |
| `POST` | `/book` | Add book manually |
| `PUT` | `/book/:id` | Update book |
| `DELETE` | `/book/:id` | Remove book |
| `PUT` | `/book/:id/monitor` | Change monitoring |

### GET /book
Query params: `?authorId=1&monitored=true&status=wanted&hasFile=false&sortKey=releaseDate`

```json
// Response 200
{
  "records": [
    {
      "id": 1,
      "authorId": 1,
      "authorName": "J.R.R. Tolkien",
      "title": "The Lord of the Rings",
      "foreignBookId": "OL27448W",
      "isbn": "0618640150",
      "isbn13": "9780618640157",
      "releaseDate": "1954-07-29",
      "overview": "One Ring to rule them all...",
      "imageUrl": "/api/v1/media/book/1/cover",
      "pageCount": 1178,
      "monitored": true,
      "status": "have",
      "editions": [...],
      "bookFile": { "id": 5, "path": "/.../LOTR.epub", "size": 2457600, "format": "epub", "quality": "epub" },
      "series": [{ "seriesId": 1, "title": "The Lord of the Rings", "position": "1-3" }],
      "addedAt": "2026-04-06T10:00:00Z"
    }
  ]
}
```

---

## Series

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/series` | List series |
| `GET` | `/series/:id` | Series detail with members |
| `PUT` | `/series/:id` | Update series |

### GET /series/:id
```json
// Response 200
{
  "id": 1,
  "foreignSeriesId": "OL123S",
  "title": "The Lord of the Rings",
  "description": "...",
  "monitored": true,
  "books": [
    { "bookId": 1, "title": "The Fellowship of the Ring", "position": "1", "monitored": true, "status": "have" },
    { "bookId": 2, "title": "The Two Towers", "position": "2", "monitored": true, "status": "wanted" },
    { "bookId": 3, "title": "The Return of the King", "position": "3", "monitored": true, "status": "wanted" }
  ]
}
```

---

## Book Files

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/bookfile` | List files |
| `GET` | `/bookfile/:id` | File detail |
| `DELETE` | `/bookfile/:id` | Remove file from disk and database |
| `PUT` | `/bookfile/:id` | Update metadata (quality, etc) |

---

## Library

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/rootfolder` | List root folders |
| `POST` | `/rootfolder` | Add root folder |
| `DELETE` | `/rootfolder/:id` | Remove root folder |
| `GET` | `/rootfolder/:id/space` | Disk space |

### POST /rootfolder
```json
// Request
{ "path": "/library", "name": "Main Library", "defaultQualityProfileId": 1 }

// Response 201
{ "id": 1, "path": "/library", "name": "Main Library", "freeSpace": 107374182400, "totalSpace": 214748364800 }
```

---

## Quality Profiles

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/qualityprofile` | List profiles |
| `GET` | `/qualityprofile/:id` | Detail |
| `POST` | `/qualityprofile` | Create profile |
| `PUT` | `/qualityprofile/:id` | Update profile |
| `DELETE` | `/qualityprofile/:id` | Remove profile |

### GET /qualityprofile/:id
```json
// Response 200
{
  "id": 1,
  "name": "eBook Quality",
  "cutoff": "epub",
  "items": [
    { "quality": "epub", "allowed": true },
    { "quality": "mobi", "allowed": true },
    { "quality": "azw3", "allowed": true },
    { "quality": "pdf", "allowed": false },
    { "quality": "unknown", "allowed": false }
  ]
}
```

---

## Search

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/search/author` | Search authors in metadata providers |
| `GET` | `/search/book` | Search books in metadata providers |
| `GET` | `/search/release/:bookId` | Search releases in indexers for a book |
| `POST` | `/search/release` | Grab: send release to download client |

### GET /search/author?term=tolkien
```json
// Response 200
[
  {
    "foreignAuthorId": "OL26320A",
    "name": "J.R.R. Tolkien",
    "overview": "...",
    "imageUrl": "https://covers.openlibrary.org/...",
    "bookCount": 35,
    "source": "openlibrary"
  }
]
```

### GET /search/release/:bookId
```json
// Response 200
[
  {
    "guid": "abc123",
    "title": "J.R.R. Tolkien - The Lord of the Rings (epub)",
    "indexer": "MyIndexer",
    "size": 2457600,
    "format": "epub",
    "seeders": 42,
    "leechers": 5,
    "approved": true,
    "rejections": [],
    "downloadUrl": "magnet:?xt=..."
  },
  {
    "guid": "def456",
    "title": "Tolkien LOTR PDF",
    "indexer": "OtherIndexer",
    "size": 15728640,
    "format": "pdf",
    "seeders": 3,
    "approved": false,
    "rejections": ["Quality not allowed in profile"]
  }
]
```

### POST /search/release
```json
// Request — grab release
{ "guid": "abc123", "indexerId": 1, "bookId": 42 }

// Response 200
{ "ok": true }
```

---

## Indexers

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/indexer` | List indexers |
| `GET` | `/indexer/:id` | Detail |
| `POST` | `/indexer` | Add indexer |
| `PUT` | `/indexer/:id` | Update indexer |
| `DELETE` | `/indexer/:id` | Remove indexer |
| `POST` | `/indexer/test` | Test connection |
| `POST` | `/indexer/testall` | Test all |

### POST /indexer
```json
// Request
{
  "name": "Prowlarr - MyIndexer",
  "type": "torznab",
  "settings": {
    "baseUrl": "http://prowlarr:9696/1/api",
    "apiKey": "abc123",
    "categories": [7000, 7020]
  },
  "enabled": true,
  "priority": 25
}
```

---

## Download Clients

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/downloadclient` | List clients |
| `GET` | `/downloadclient/:id` | Detail |
| `POST` | `/downloadclient` | Add client |
| `PUT` | `/downloadclient/:id` | Update client |
| `DELETE` | `/downloadclient/:id` | Remove client |
| `POST` | `/downloadclient/test` | Test connection |

### POST /downloadclient
```json
// Request
{
  "name": "qBittorrent",
  "type": "qbittorrent",
  "settings": {
    "host": "qbittorrent",
    "port": 8080,
    "username": "admin",
    "password": "secret",
    "category": "books"
  },
  "enabled": true,
  "priority": 1
}
```

---

## Download Queue

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/queue` | List downloads in progress |
| `GET` | `/queue/:id` | Detail |
| `DELETE` | `/queue/:id` | Remove from queue (cancel download) |

### GET /queue
```json
// Response 200
{
  "records": [
    {
      "id": 1,
      "bookId": 42,
      "bookTitle": "The Lord of the Rings",
      "authorName": "J.R.R. Tolkien",
      "quality": "epub",
      "size": 2457600,
      "status": "downloading",
      "progress": 65.4,
      "downloadClient": "qBittorrent",
      "indexer": "MyIndexer",
      "addedAt": "2026-04-06T12:00:00Z",
      "estimatedCompletionTime": "2026-04-06T12:05:00Z"
    }
  ]
}
```

---

## Notifications

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/notification` | List notifications |
| `POST` | `/notification` | Add notification |
| `PUT` | `/notification/:id` | Update |
| `DELETE` | `/notification/:id` | Remove |
| `POST` | `/notification/test` | Test delivery |

### POST /notification
```json
// Request — webhook
{
  "name": "My Webhook",
  "type": "webhook",
  "settings": { "url": "https://example.com/hook", "method": "POST" },
  "onGrab": true,
  "onDownload": true,
  "onUpgrade": false
}

// Request — discord
{
  "name": "Discord",
  "type": "discord",
  "settings": { "webhookUrl": "https://discord.com/api/webhooks/..." },
  "onGrab": true,
  "onDownload": true
}

// Request — email
{
  "name": "Email",
  "type": "email",
  "settings": {
    "server": "smtp.gmail.com",
    "port": 587,
    "username": "user@gmail.com",
    "password": "app-password",
    "from": "bookaneer@example.com",
    "to": ["user@example.com"],
    "useTls": true
  },
  "onGrab": false,
  "onDownload": true
}
```

---

## History

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/history` | List history |
| `GET` | `/history/author/:authorId` | History by author |
| `GET` | `/history/book/:bookId` | History by book |

### GET /history
Query params: `?eventType=grabbed&page=1&pageSize=20`

Event types: `grabbed`, `downloadCompleted`, `downloadFailed`, `bookFileDeleted`, `bookFileRenamed`, `bookImported`

```json
// Response 200
{
  "records": [
    {
      "id": 1,
      "bookId": 42,
      "bookTitle": "The Lord of the Rings",
      "authorName": "J.R.R. Tolkien",
      "eventType": "grabbed",
      "sourceTitle": "Tolkien - LOTR (epub)",
      "quality": "epub",
      "date": "2026-04-06T12:00:00Z",
      "data": { "indexer": "MyIndexer", "downloadClient": "qBittorrent" }
    }
  ]
}
```

---

## Wanted

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/wanted/missing` | Monitored books without files |
| `GET` | `/wanted/cutoff` | Books with quality below cutoff |

### GET /wanted/missing
```json
// Response 200 (same structure as GET /book, filtered)
{
  "page": 1, "pageSize": 20, "totalRecords": 5,
  "records": [
    { "id": 2, "title": "The Two Towers", "authorName": "J.R.R. Tolkien", "monitored": true, "status": "wanted" }
  ]
}
```

---

## Reader

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/reader/:bookFileId` | Book metadata for the reader |
| `GET` | `/reader/:bookFileId/content` | Serve the EPUB file (stream) |
| `GET` | `/reader/:bookFileId/progress` | Current reading position |
| `PUT` | `/reader/:bookFileId/progress` | Save reading position |

### GET /reader/:bookFileId
```json
// Response 200
{
  "bookFileId": 5,
  "bookId": 1,
  "title": "The Lord of the Rings",
  "authorName": "J.R.R. Tolkien",
  "format": "epub",
  "size": 2457600,
  "contentUrl": "/api/v1/reader/5/content",
  "progress": { "position": "epubcfi(/6/4[chap01]!/4/2/1:0)", "percentage": 23.5, "updatedAt": "2026-04-06T14:00:00Z" }
}
```

### PUT /reader/:bookFileId/progress
```json
// Request
{ "position": "epubcfi(/6/8[chap03]!/4/2/1:0)", "percentage": 45.2 }
```

---

## OPDS

Not part of the REST API v1. Served directly:

| Path | Description |
|------|-------------|
| `GET /opds` | Root catalog |
| `GET /opds/authors` | Browse by author |
| `GET /opds/authors/:id` | Books by an author |
| `GET /opds/series` | Browse by series |
| `GET /opds/series/:id` | Books in a series |
| `GET /opds/recent` | Recently added |
| `GET /opds/search?q=...` | Search |

Format: Atom XML (OPDS 1.2). Auth via HTTP Basic (username + api key as password).

---

## Commands

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/command` | List commands (running + recent) |
| `GET` | `/command/:id` | Command status |
| `POST` | `/command` | Execute command |

### POST /command
```json
// Search releases for a book
{ "name": "BookSearch", "bookIds": [42] }

// Search all wanted books
{ "name": "MissingBookSearch" }

// Library scan
{ "name": "LibraryScan", "path": "/library" }

// RSS sync
{ "name": "RssSync" }

// Rename files for an author
{ "name": "RenameFiles", "authorId": 1 }

// Backup
{ "name": "Backup" }

// Response 201
{ "id": "01HYX...", "name": "BookSearch", "status": "queued", "queuedAt": "2026-04-06T12:00:00Z" }
```

---

## Config

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/config/general` | General config (port, auth, etc) |
| `PUT` | `/config/general` | Update config |
| `GET` | `/config/naming` | Renaming config |
| `PUT` | `/config/naming` | Update naming |

### GET /config/naming
```json
// Response 200
{
  "renamingEnabled": true,
  "authorFolderFormat": "$Author",
  "bookFileFormat": "$Author - $Title{ ($SeriesName #$SeriesPosition)}",
  "replaceSpaces": false,
  "colonReplacement": "dash"
}
```

---

## Media (covers)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/media/author/:id/cover` | Author cover art |
| `GET` | `/media/book/:id/cover` | Book cover art |

Returns image directly (Content-Type: image/jpeg). Cached on disk.

---

## System

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/system/status` | System status (version, uptime, etc) |
| `GET` | `/system/health` | Health checks |
| `GET` | `/system/task` | Scheduled tasks and status |
| `GET` | `/system/backup` | List backups |
| `POST` | `/system/backup` | Create backup |
| `GET` | `/system/log` | Recent logs |

### GET /system/status
```json
// Response 200
{
  "version": "0.1.0",
  "buildTime": "2026-04-06T10:00:00Z",
  "osName": "linux",
  "osArch": "arm64",
  "runtimeVersion": "go1.26.0",
  "startTime": "2026-04-06T09:00:00Z",
  "appDataDir": "/data",
  "libraryDir": "/library"
}
```

### GET /system/health
```json
// Response 200
[
  { "type": "ok", "message": "All indexers are available" },
  { "type": "warning", "message": "No download client configured" },
  { "type": "error", "message": "Root folder /library is not accessible" }
]
```

---

## WebSocket

Endpoint: `ws://host:9090/api/v1/ws`

Auth: `?apikey=xxx` in the connection URL.

Messages are JSON:

```json
// Command status update
{ "event": "command", "data": { "id": "01HYX...", "name": "BookSearch", "status": "completed" } }

// Book status changed
{ "event": "book", "data": { "id": 42, "status": "have" } }

// Download progress
{ "event": "queue", "data": { "id": 1, "bookId": 42, "progress": 65.4 } }

// Health check update
{ "event": "health", "data": [{ "type": "ok", "message": "..." }] }
```

---

## Tags

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/tag` | List tags |
| `POST` | `/tag` | Create tag |
| `PUT` | `/tag/:id` | Update tag |
| `DELETE` | `/tag/:id` | Remove tag |
