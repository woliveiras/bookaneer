import type { Book } from "./book"

export type CommandStatus = "queued" | "running" | "completed" | "failed" | "cancelled"

export interface ActiveCommand {
  id: string
  name: string
  status: CommandStatus
  priority: number
  payload?: Record<string, unknown>
  result?: Record<string, unknown>
  trigger: string
  queuedAt: string
  startedAt?: string
  endedAt?: string
}

export interface WantedResponse {
  page: number
  pageSize: number
  totalRecords: number
  sortKey: string
  sortDirection: string
  records: Book[]
}

export interface SearchCommandResponse {
  commandId: string
  message: string
}

export interface GrabResult {
  bookId: number
  title: string
  source: string
  providerName: string
  format: string
  size: number
  downloadId: string
  clientName: string
}

export interface BookSearchResult {
  title: string
  source: string
  provider: string
  format: string
  size: number
  downloadUrl: string
  seeders?: number
}

export interface BookSearchResponse {
  results: BookSearchResult[]
  noResults: boolean
}

export type HistoryEventType =
  | "grabbed"
  | "downloadCompleted"
  | "downloadFailed"
  | "bookFileDeleted"
  | "bookFileRenamed"
  | "bookImported"
  | "contentMismatch"
  | "wrongContent"
  | "metadataExtracted"

export interface HistoryItem {
  id: number
  bookId?: number
  authorId?: number
  eventType: HistoryEventType
  sourceTitle: string
  quality: string
  data: Record<string, unknown>
  date: string
  bookTitle?: string
  authorName?: string
}

export interface BlocklistItem {
  id: number
  bookId: number
  sourceTitle: string
  quality: string
  reason: string
  date: string
  bookTitle: string
  authorName: string
}
