import * as z from "zod"

export const CommandStatusSchema = z.enum(["queued", "running", "completed", "failed", "cancelled"])

export const ActiveCommandSchema = z.object({
  id: z.string(),
  name: z.string(),
  status: CommandStatusSchema,
  priority: z.number(),
  payload: z.record(z.string(), z.unknown()).optional(),
  result: z.record(z.string(), z.unknown()).optional(),
  trigger: z.string(),
  queuedAt: z.string(),
  startedAt: z.string().optional(),
  endedAt: z.string().optional(),
})

export const GrabResultSchema = z.object({
  bookId: z.number(),
  title: z.string(),
  source: z.string(),
  providerName: z.string(),
  format: z.string(),
  size: z.number(),
  downloadId: z.string(),
  clientName: z.string(),
})

export const BookSearchResultSchema = z.object({
  title: z.string(),
  source: z.string(),
  provider: z.string(),
  format: z.string(),
  size: z.number(),
  downloadUrl: z.string(),
  seeders: z.number().optional(),
})

export const BookSearchResponseSchema = z.object({
  results: z.array(BookSearchResultSchema),
  noResults: z.boolean(),
})

export const HistoryEventTypeSchema = z.enum([
  "grabbed",
  "downloadCompleted",
  "downloadFailed",
  "bookFileDeleted",
  "bookFileRenamed",
  "bookImported",
  "contentMismatch",
  "wrongContent",
  "metadataExtracted",
])

export const HistoryItemSchema = z.object({
  id: z.number(),
  bookId: z.number().optional(),
  authorId: z.number().optional(),
  eventType: HistoryEventTypeSchema,
  sourceTitle: z.string(),
  quality: z.string(),
  data: z.record(z.string(), z.unknown()),
  date: z.string(),
  bookTitle: z.string().optional(),
  authorName: z.string().optional(),
})

export const BlocklistItemSchema = z.object({
  id: z.number(),
  bookId: z.number(),
  sourceTitle: z.string(),
  quality: z.string(),
  reason: z.string(),
  date: z.string(),
  bookTitle: z.string(),
  authorName: z.string(),
})

export type CommandStatus = z.infer<typeof CommandStatusSchema>
export type ActiveCommand = z.infer<typeof ActiveCommandSchema>
export type GrabResult = z.infer<typeof GrabResultSchema>
export type BookSearchResult = z.infer<typeof BookSearchResultSchema>
export type BookSearchResponse = z.infer<typeof BookSearchResponseSchema>
export type HistoryEventType = z.infer<typeof HistoryEventTypeSchema>
export type HistoryItem = z.infer<typeof HistoryItemSchema>
export type BlocklistItem = z.infer<typeof BlocklistItemSchema>
