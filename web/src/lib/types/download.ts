export type DownloadClientType = "sabnzbd" | "qbittorrent" | "transmission" | "blackhole"

export interface DownloadClient {
  id: number
  name: string
  type: DownloadClientType
  host: string
  port: number
  useTls: boolean
  username: string
  password: string
  apiKey: string
  category: string
  recentPriority: number
  olderPriority: number
  removeCompletedAfter: number
  enabled: boolean
  priority: number
  nzbFolder: string
  torrentFolder: string
  watchFolder: string
  createdAt: string
  updatedAt: string
}

export interface CreateDownloadClientInput {
  name: string
  type: DownloadClientType
  host?: string
  port?: number
  useTls?: boolean
  username?: string
  password?: string
  apiKey?: string
  category?: string
  recentPriority?: number
  olderPriority?: number
  removeCompletedAfter?: number
  enabled?: boolean
  priority?: number
  nzbFolder?: string
  torrentFolder?: string
  watchFolder?: string
}

export type DownloadStatus = "queued" | "downloading" | "paused" | "completed" | "seeding" | "failed" | "extracted" | "processing"

export interface QueueItem {
  id: number
  bookId: number
  downloadClientId?: number
  indexerId?: number
  externalId: string
  title: string
  size: number
  format: string
  status: DownloadStatus
  progress: number
  downloadUrl: string
  addedAt: string
  bookTitle: string
  clientName: string
}

export type GrabStatus = "pending" | "sent" | "downloading" | "completed" | "failed" | "imported"

export interface Grab {
  id: number
  bookId: number
  indexerId: number
  releaseTitle: string
  downloadUrl: string
  size: number
  quality: string
  clientId: number
  downloadId: string
  status: GrabStatus
  errorMessage: string
  grabbedAt: string
  completedAt?: string
}

export interface CreateGrabInput {
  bookId: number
  indexerId: number
  releaseTitle: string
  downloadUrl: string
  size?: number
  quality?: string
  clientId: number
}

export interface TestDownloadClientResponse {
  success: boolean
  message: string
}
