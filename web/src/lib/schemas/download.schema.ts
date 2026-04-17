import * as z from "zod"

export const DownloadClientTypeSchema = z.enum([
  "sabnzbd",
  "qbittorrent",
  "transmission",
  "blackhole",
])

export const DownloadClientSchema = z.object({
  id: z.number(),
  name: z.string(),
  type: DownloadClientTypeSchema,
  host: z.string(),
  port: z.number(),
  useTls: z.boolean(),
  username: z.string(),
  password: z.string(),
  apiKey: z.string(),
  category: z.string(),
  recentPriority: z.number(),
  olderPriority: z.number(),
  removeCompletedAfter: z.number(),
  enabled: z.boolean(),
  priority: z.number(),
  nzbFolder: z.string(),
  torrentFolder: z.string(),
  watchFolder: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const CreateDownloadClientInputSchema = z.object({
  name: z.string(),
  type: DownloadClientTypeSchema,
  host: z.string().optional(),
  port: z.number().optional(),
  useTls: z.boolean().optional(),
  username: z.string().optional(),
  password: z.string().optional(),
  apiKey: z.string().optional(),
  category: z.string().optional(),
  recentPriority: z.number().optional(),
  olderPriority: z.number().optional(),
  removeCompletedAfter: z.number().optional(),
  enabled: z.boolean().optional(),
  priority: z.number().optional(),
  nzbFolder: z.string().optional(),
  torrentFolder: z.string().optional(),
  watchFolder: z.string().optional(),
})

export const DownloadStatusSchema = z.enum([
  "queued",
  "downloading",
  "paused",
  "completed",
  "seeding",
  "failed",
  "extracted",
  "processing",
])

export const QueueItemSchema = z.object({
  id: z.number(),
  bookId: z.number(),
  downloadClientId: z.number().optional(),
  indexerId: z.number().optional(),
  externalId: z.string(),
  title: z.string(),
  size: z.number(),
  format: z.string(),
  status: DownloadStatusSchema,
  progress: z.number(),
  downloadUrl: z.string(),
  addedAt: z.string(),
  bookTitle: z.string(),
  clientName: z.string(),
})

export const GrabStatusSchema = z.enum([
  "pending",
  "sent",
  "downloading",
  "completed",
  "failed",
  "imported",
])

export const GrabSchema = z.object({
  id: z.number(),
  bookId: z.number(),
  indexerId: z.number(),
  releaseTitle: z.string(),
  downloadUrl: z.string(),
  size: z.number(),
  quality: z.string(),
  clientId: z.number(),
  downloadId: z.string(),
  status: GrabStatusSchema,
  errorMessage: z.string(),
  grabbedAt: z.string(),
  completedAt: z.string().optional(),
})

export const CreateGrabInputSchema = z.object({
  bookId: z.number(),
  indexerId: z.number(),
  releaseTitle: z.string(),
  downloadUrl: z.string(),
  size: z.number().optional(),
  quality: z.string().optional(),
  clientId: z.number(),
})

export const TestDownloadClientResponseSchema = z.object({
  success: z.boolean(),
  message: z.string(),
})

export type DownloadClientType = z.infer<typeof DownloadClientTypeSchema>
export type DownloadClient = z.infer<typeof DownloadClientSchema>
export type CreateDownloadClientInput = z.infer<typeof CreateDownloadClientInputSchema>
export type DownloadStatus = z.infer<typeof DownloadStatusSchema>
export type QueueItem = z.infer<typeof QueueItemSchema>
export type GrabStatus = z.infer<typeof GrabStatusSchema>
export type Grab = z.infer<typeof GrabSchema>
export type CreateGrabInput = z.infer<typeof CreateGrabInputSchema>
export type TestDownloadClientResponse = z.infer<typeof TestDownloadClientResponseSchema>
