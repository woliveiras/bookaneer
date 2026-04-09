import type { DownloadClientType } from "../../lib/api"

export interface DownloadClientPreset {
  id: string
  name: string
  type: DownloadClientType
  host: string
  port: number
  description: string
}

export const DOWNLOAD_CLIENT_PRESETS: {
  usenet: DownloadClientPreset[]
  torrents: DownloadClientPreset[]
} = {
  usenet: [
    {
      id: "sabnzbd",
      name: "SABnzbd",
      type: "sabnzbd",
      host: "localhost",
      port: 8080,
      description: "Popular Usenet downloader",
    },
  ],
  torrents: [
    {
      id: "qbittorrent",
      name: "qBittorrent",
      type: "qbittorrent",
      host: "localhost",
      port: 8080,
      description: "Feature-rich torrent client",
    },
    {
      id: "transmission",
      name: "Transmission",
      type: "transmission",
      host: "localhost",
      port: 9091,
      description: "Lightweight torrent client",
    },
    {
      id: "blackhole-nzb",
      name: "Blackhole (NZB)",
      type: "blackhole",
      host: "",
      port: 0,
      description: "Drop NZB files to folder",
    },
    {
      id: "blackhole-torrent",
      name: "Blackhole (Torrent)",
      type: "blackhole",
      host: "",
      port: 0,
      description: "Drop torrent files to folder",
    },
  ],
}
