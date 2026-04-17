import * as z from "zod"

export const AuthorSchema = z.object({
  id: z.number(),
  name: z.string(),
  sortName: z.string(),
  foreignId: z.string(),
  overview: z.string(),
  imageUrl: z.string(),
  status: z.string(),
  monitored: z.boolean(),
  path: z.string(),
  addedAt: z.string(),
  updatedAt: z.string(),
  bookCount: z.number().optional(),
  bookFileCount: z.number().optional(),
})

export const AuthorStatsSchema = z.object({
  bookCount: z.number(),
  bookFileCount: z.number(),
  missingBooks: z.number(),
  totalSizeBytes: z.number(),
})

export const CreateAuthorInputSchema = z.object({
  name: z.string(),
  sortName: z.string().optional(),
  foreignId: z.string().optional(),
  overview: z.string().optional(),
  imageUrl: z.string().optional(),
  status: z.string().optional(),
  monitored: z.boolean().optional(),
  path: z.string().optional(),
})

export const UpdateAuthorInputSchema = CreateAuthorInputSchema.partial()

export const ListAuthorsParamsSchema = z.object({
  monitored: z.boolean().optional(),
  status: z.string().optional(),
  search: z.string().optional(),
  sortBy: z.string().optional(),
  sortDir: z.string().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
})

export type Author = z.infer<typeof AuthorSchema>
export type AuthorStats = z.infer<typeof AuthorStatsSchema>
export type CreateAuthorInput = z.infer<typeof CreateAuthorInputSchema>
export type UpdateAuthorInput = z.infer<typeof UpdateAuthorInputSchema>
export type ListAuthorsParams = z.infer<typeof ListAuthorsParamsSchema>
