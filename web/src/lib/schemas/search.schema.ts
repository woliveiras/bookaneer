import * as z from "zod"

export const MetadataAuthorResultSchema = z.object({
  foreignId: z.string(),
  name: z.string(),
  birthYear: z.number().optional(),
  deathYear: z.number().optional(),
  photoUrl: z.string().optional(),
  worksCount: z.number().optional(),
  provider: z.string(),
})

export const MetadataBookResultSchema = z.object({
  foreignId: z.string(),
  title: z.string(),
  authors: z.array(z.string()).optional(),
  publishedYear: z.number().optional(),
  coverUrl: z.string().optional(),
  isbn10: z.string().optional(),
  isbn13: z.string().optional(),
  provider: z.string(),
})

const MetadataLinkSchema = z.object({ type: z.string(), url: z.string() })

export const MetadataAuthorSchema = z.object({
  foreignId: z.string(),
  name: z.string(),
  sortName: z.string().optional(),
  bio: z.string().optional(),
  birthDate: z.string().optional(),
  deathDate: z.string().optional(),
  photoUrl: z.string().optional(),
  website: z.string().optional(),
  wikipedia: z.string().optional(),
  nationality: z.string().optional(),
  provider: z.string(),
  links: z.array(MetadataLinkSchema).optional(),
})

export const MetadataBookSchema = z.object({
  foreignId: z.string(),
  title: z.string(),
  subtitle: z.string().optional(),
  authors: z.array(z.string()).optional(),
  authorIds: z.array(z.string()).optional(),
  description: z.string().optional(),
  publishedDate: z.string().optional(),
  publisher: z.string().optional(),
  pageCount: z.number().optional(),
  language: z.string().optional(),
  isbn10: z.string().optional(),
  isbn13: z.string().optional(),
  asin: z.string().optional(),
  coverUrl: z.string().optional(),
  genres: z.array(z.string()).optional(),
  subjects: z.array(z.string()).optional(),
  series: z.string().optional(),
  seriesPosition: z.number().optional(),
  averageRating: z.number().optional(),
  ratingsCount: z.number().optional(),
  provider: z.string(),
  links: z.array(MetadataLinkSchema).optional(),
})

export function MetadataSearchResponseSchema<T extends z.ZodTypeAny>(itemSchema: T) {
  return z.object({ results: z.array(itemSchema), total: z.number() })
}

export const ReleaseSourceTypeSchema = z.enum(["library", "indexer", "both"])

export const DigitalLibraryResultSchema = z.object({
  id: z.string(),
  title: z.string(),
  authors: z.array(z.string()).optional(),
  publisher: z.string().optional(),
  year: z.number().optional(),
  language: z.string().optional(),
  format: z.string(),
  size: z.number(),
  isbn: z.string().optional(),
  coverUrl: z.string().optional(),
  downloadUrl: z.string().optional(),
  infoUrl: z.string().optional(),
  provider: z.string(),
  score: z.number().optional(),
})

// Server-driven column configuration — mirrors the backend's ColumnConfig shape
export const ColumnRenderTypeSchema = z.enum([
  "text",
  "badge",
  "size",
  "number",
  "peers",
  "indexer",
])

export const ColumnAlignSchema = z.enum(["left", "center", "right"])

export const ColumnColorHintSchema = z.object({
  type: z.enum(["map", "static"]),
  value: z.string(),
})

export const ColumnSchemaSchema = z.object({
  key: z.string(),
  label: z.string(),
  renderType: ColumnRenderTypeSchema,
  align: ColumnAlignSchema,
  width: z.string(),
  hideMobile: z.boolean().optional(),
  colorHint: ColumnColorHintSchema.optional(),
  fallback: z.string().optional(),
  uppercase: z.boolean().optional(),
  sortable: z.boolean().optional(),
  sortKey: z.string().optional(),
})

export const ColumnConfigSchema = z.object({
  columns: z.array(ColumnSchemaSchema),
  gridTemplate: z.string(),
  supportedFilters: z.array(z.string()).optional(),
})

export const DigitalLibrarySearchResponseSchema = z.object({
  results: z.array(DigitalLibraryResultSchema),
  total: z.number().optional(),
  columnConfig: ColumnConfigSchema.optional(),
})

export const UnifiedReleaseSchema = z.object({
  id: z.string(),
  title: z.string(),
  authors: z.array(z.string()).optional(),
  format: z.string().optional(),
  size: z.number(),
  downloadUrl: z.string(),
  infoUrl: z.string().optional(),
  provider: z.string(),
  sourceType: ReleaseSourceTypeSchema,
  language: z.string().optional(),
  year: z.number().optional(),
  isbn: z.string().optional(),
  coverUrl: z.string().optional(),
  score: z.number().optional(),
  seeders: z.number().optional(),
  leechers: z.number().optional(),
  grabs: z.number().optional(),
  indexerId: z.number().optional(),
  indexerName: z.string().optional(),
  quality: z.string().optional(),
})

export type MetadataAuthorResult = z.infer<typeof MetadataAuthorResultSchema>
export type MetadataBookResult = z.infer<typeof MetadataBookResultSchema>
export type MetadataAuthor = z.infer<typeof MetadataAuthorSchema>
export type MetadataBook = z.infer<typeof MetadataBookSchema>
export type MetadataSearchResponse<T> = { results: T[]; total: number }
export type ReleaseSourceType = z.infer<typeof ReleaseSourceTypeSchema>
export type DigitalLibraryResult = z.infer<typeof DigitalLibraryResultSchema>
export type ColumnRenderType = z.infer<typeof ColumnRenderTypeSchema>
export type ColumnAlign = z.infer<typeof ColumnAlignSchema>
export type ColumnColorHint = z.infer<typeof ColumnColorHintSchema>
export type ColumnSchema = z.infer<typeof ColumnSchemaSchema>
export type ColumnConfig = z.infer<typeof ColumnConfigSchema>
export type DigitalLibrarySearchResponse = z.infer<typeof DigitalLibrarySearchResponseSchema>
export type UnifiedRelease = z.infer<typeof UnifiedReleaseSchema>
