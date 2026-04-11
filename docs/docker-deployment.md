# Docker Deployment Guide

How to run Bookaneer on your local network using Docker.

## Requirements

- Docker Engine 20.10+ (or Docker Desktop)
- A folder with your ebook collection (EPUB, PDF, etc.)

## Quick Start

### Option 1: Docker Compose (recommended)

1. Create a directory for Bookaneer:

```sh
mkdir bookaneer && cd bookaneer
```

2. Create a `docker-compose.yml`:

```yaml
services:
  bookaneer:
    image: ghcr.io/woliveiras/bookaneer:latest
    container_name: bookaneer
    ports:
      - "9090:9090"
    environment:
      - BOOKANEER_LOG_LEVEL=info
    volumes:
      - bookaneer_data:/data
      - /path/to/your/books:/library
    restart: unless-stopped

volumes:
  bookaneer_data:
```

3. Replace `/path/to/your/books` with the actual path to your ebook folder.

4. Start the service:

```sh
docker compose up -d
```

5. Open `http://localhost:9090` in your browser.

### Option 2: Docker Run

```sh
docker run -d \
  --name bookaneer \
  -p 9090:9090 \
  -v bookaneer_data:/data \
  -v /path/to/your/books:/library \
  --restart unless-stopped \
  ghcr.io/woliveiras/bookaneer:latest
```

## First Login

On first boot, Bookaneer creates an admin account:

- **Username:** `admin`  
- **Password:** generated automatically and printed in the container logs

To see the password:

```sh
docker logs bookaneer 2>&1 | grep "Password:"
```

The credentials are also saved to `/data/admin_credentials.txt` inside the container volume. Change the password after your first login.

To set a custom password on first boot, add this environment variable:

```yaml
environment:
  - BOOKANEER_ADMIN_PASSWORD=your-secure-password
```

## Accessing from Other Devices

Bookaneer binds to `0.0.0.0` by default, so any device on your local network can access it.

1. Find your machine's IP address:

```sh
# macOS
ipconfig getifaddr en0

# Linux
hostname -I | awk '{print $1}'

# Windows (PowerShell)
(Get-NetIPAddress -AddressFamily IPv4 -InterfaceAlias "Wi-Fi").IPAddress
```

2. From another device, open: `http://<your-ip>:9090`

Example: if your IP is `192.168.1.100`, open `http://192.168.1.100:9090`.

## Volumes

| Container Path | Purpose |
|---|---|
| `/data` | Database, configuration, logs, credentials. **Back this up.** |
| `/library` | Your ebook collection. Bookaneer reads and writes files here. |

Use a named volume for `/data` (keeps the database safe across container recreations). Use a bind mount for `/library` (maps to your existing ebook folder).

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `BOOKANEER_PORT` | `9090` | HTTP server port |
| `BOOKANEER_BIND_ADDRESS` | `0.0.0.0` | Listen address |
| `BOOKANEER_DATA_DIR` | `/data` | Data directory inside container |
| `BOOKANEER_LIBRARY_DIR` | `/library` | Library directory inside container |
| `BOOKANEER_LOG_LEVEL` | `info` | Log verbosity: `debug`, `info`, `warn`, `error` |
| `BOOKANEER_ADMIN_PASSWORD` | *(generated)* | Admin password on first boot only |

## Changing the Port

To run on a different port (e.g., 8080), change the port mapping and the environment variable:

```yaml
services:
  bookaneer:
    ports:
      - "8080:8080"
    environment:
      - BOOKANEER_PORT=8080
```

## OPDS — Access from Reading Apps

Bookaneer includes an OPDS catalog, which lets you browse and download books from reading apps like KOReader, Moon+ Reader, Librera, or Calibre.

In your reading app, add an OPDS catalog with the URL:

```
http://<your-ip>:9090/opds
```

## Backups

The database and all configuration live in the `/data` volume. To back up:

```sh
# If using a named volume
docker run --rm -v bookaneer_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/bookaneer-backup.tar.gz -C /data .

# If using a bind mount
tar czf bookaneer-backup.tar.gz -C ./data .
```

## Updating

```sh
docker compose pull
docker compose up -d
```

Or with `docker run`:

```sh
docker pull ghcr.io/woliveiras/bookaneer:latest
docker stop bookaneer
docker rm bookaneer
# Then re-run the docker run command from above
```

Your data is preserved in the volume.

## Troubleshooting

### Port already in use

Change the host port in the port mapping: `-p 8080:9090` (access on 8080, container still listens on 9090).

### Permission denied on library folder

Make sure the library folder is readable by the container. On Linux:

```sh
chmod -R a+r /path/to/your/books
```

### Container is unhealthy

Check the logs:

```sh
docker logs bookaneer
```

The container includes a built-in health check that runs every 30 seconds. If it reports `unhealthy`, the logs will show what went wrong.

### Can't access from another device

- Verify both devices are on the same network
- Check your firewall allows port 9090
- On macOS: System Settings → Network → Firewall → allow incoming connections
- On Linux: `sudo ufw allow 9090/tcp`
