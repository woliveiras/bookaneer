import * as z from "zod"

export const BookSchema = z.object({
  id: z.number(),
  authorId: z.number(),
  title: z.string(),
  sortTitle: z.string(),
  foreignId: z.string(),
  isbn: z.string(),
  isbn13: z.string(),
  releaseDate: z.string(),
  overview: z.string(),
  imageUrl: z.string(),
  pageCount: z.number(),
  userRating: z.number().optional(),
  inWishlist: z.boolean(),
  addedAt: z.string(),
  updatedAt: z.string(),
  authorName: z.string().optional(),
  hasFile: z.boolean().optional(),
  fileFormat: z.string().optional(),
})

export const EditionSchema = z.object({
  id: z.number(),
  bookId: z.number(),
  foreignId: z.string(),
  title: z.string(),
  isbn: z.string(),
  isbn13: z.string(),
  format: z.string(),
  publisher: z.string(),
  releaseDate: z.string(),
  pageCount: z.number(),
  language: z.string(),
})

export const BookFileSchema = z.object({
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
  contentMismatch: z.boolean(),
})

export const BookWithEditionsSchema = BookSchema.extend({
  editions: z.array(EditionSchema),
  files: z.array(BookFileSchema).optional(),
})

export const CreateBookInputSchema = z.object({
  authorId: z.number(),
  title: z.string(),
  sortTitle: z.string().optional(),
  foreignId: z.string().optional(),
  isbn: z.string().optional(),
  isbn13: z.string().optional(),
  releaseDate: z.string().optional(),
  overview: z.string().optional(),
  imageUrl: z.string().optional(),
  pageCount: z.number().optional(),
  userRating: z.number().optional(),
  inWishlist: z.boolean().optional(),
})

export const ListBooksParamsSchema = z.object({
  authorId: z.number().optional(),
  missing: z.boolean().optional(),
  inWishlist: z.boolean().optional(),
  search: z.string().optional(),
  sortBy: z.string().optional(),
  sortDir: z.string().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
})

export function PaginatedResponseSchema<T extends z.ZodTypeAny>(itemSchema: T) {
  return z.object({
    records: z.array(itemSchema),
    totalRecords: z.number(),
  })
}

export type Book = z.infer<typeof BookSchema>
export type Edition = z.infer<typeof EditionSchema>
export type BookFile = z.infer<typeof BookFileSchema>
export type BookWithEditions = z.infer<typeof BookWithEditionsSchema>
export type CreateBookInput = z.infer<typeof CreateBookInputSchema>
export type ListBooksParams = z.infer<typeof ListBooksParamsSchema>
export type PaginatedResponse<T> = { records: T[]; totalRecords: number }
