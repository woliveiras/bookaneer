import * as z from "zod"

export const ReaderBookFileSchema = z.object({
  id: z.number(),
  bookId: z.number(),
  editionId: z.number().optional(),
  path: z.string(),
  relativePath: z.string(),
  size: z.number(),
  format: z.string(),
  quality: z.string(),
  hash: z.string(),
  addedAt: z.string(),
  bookTitle: z.string().optional(),
  authorName: z.string().optional(),
  coverUrl: z.string().optional(),
})

export const ReadingProgressSchema = z.object({
  id: z.number().optional(),
  bookFileId: z.number(),
  userId: z.number().optional(),
  position: z.string(),
  percentage: z.number(),
  updatedAt: z.string().optional(),
})

export const SaveProgressInputSchema = z.object({
  position: z.string(),
  percentage: z.number(),
})

export const BookmarkSchema = z.object({
  id: z.number(),
  bookFileId: z.number(),
  userId: z.number(),
  position: z.string(),
  title: z.string(),
  note: z.string(),
  createdAt: z.string(),
})

export const CreateBookmarkInputSchema = z.object({
  position: z.string(),
  title: z.string(),
  note: z.string().optional(),
})

export type ReaderBookFile = z.infer<typeof ReaderBookFileSchema>
export type ReadingProgress = z.infer<typeof ReadingProgressSchema>
export type SaveProgressInput = z.infer<typeof SaveProgressInputSchema>
export type Bookmark = z.infer<typeof BookmarkSchema>
export type CreateBookmarkInput = z.infer<typeof CreateBookmarkInputSchema>
