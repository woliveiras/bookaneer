import * as z from "zod"
import { ColumnConfigSchema } from "./search.schema"

export const IndexerTypeSchema = z.enum(["newznab", "torznab"])

export const IndexerSchema = z.object({
  id: z.number(),
  name: z.string(),
  type: IndexerTypeSchema,
  baseUrl: z.string(),
  apiPath: z.string(),
  apiKey: z.string(),
  categories: z.string(),
  priority: z.number(),
  enabled: z.boolean(),
  enableRss: z.boolean(),
  enableAutomaticSearch: z.boolean(),
  enableInteractiveSearch: z.boolean(),
  additionalParameters: z.string(),
  minimumSeeders: z.number(),
  seedRatio: z.number().nullable().optional(),
  seedTime: z.number().nullable().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export const CreateIndexerInputSchema = z.object({
  name: z.string(),
  type: IndexerTypeSchema,
  baseUrl: z.string(),
  apiPath: z.string().optional(),
  apiKey: z.string(),
  categories: z.string().optional(),
  priority: z.number().optional(),
  enabled: z.boolean().optional(),
  enableRss: z.boolean().optional(),
  enableAutomaticSearch: z.boolean().optional(),
  enableInteractiveSearch: z.boolean().optional(),
  additionalParameters: z.string().optional(),
  minimumSeeders: z.number().optional(),
  seedRatio: z.number().nullable().optional(),
  seedTime: z.number().nullable().optional(),
})

export const UpdateIndexerInputSchema = CreateIndexerInputSchema.extend({
  id: z.number().optional(),
})

export const SearchResultSchema = z.object({
  guid: z.string(),
  title: z.string(),
  description: z.string().optional(),
  size: z.number(),
  pubDate: z.string(),
  category: z.string().optional(),
  categoryId: z.string().optional(),
  downloadUrl: z.string(),
  infoUrl: z.string().optional(),
  comments: z.number().optional(),
  seeders: z.number().optional(),
  leechers: z.number().optional(),
  grabs: z.number().optional(),
  quality: z.string().optional(),
  qualityRank: z.number().optional(),
  indexerId: z.number(),
  indexerName: z.string(),
})

export const SearchResponseSchema = z.object({
  results: z.array(SearchResultSchema),
  total: z.number(),
  columnConfig: ColumnConfigSchema.optional(),
})

export const TestIndexerResponseSchema = z.object({
  success: z.boolean(),
  message: z.string(),
})

export const IndexerOptionsSchema = z.object({
  minimumAge: z.number(),
  retention: z.number(),
  maximumSize: z.number(),
  rssSyncInterval: z.number(),
  preferIndexerFlags: z.boolean(),
  availabilityDelay: z.number(),
  updatedAt: z.string(),
})

export const UpdateIndexerOptionsInputSchema = IndexerOptionsSchema.omit({ updatedAt: true })

export type Indexer = z.infer<typeof IndexerSchema>
export type CreateIndexerInput = z.infer<typeof CreateIndexerInputSchema>
export type UpdateIndexerInput = z.infer<typeof UpdateIndexerInputSchema>
export type SearchResult = z.infer<typeof SearchResultSchema>
export type SearchResponse = z.infer<typeof SearchResponseSchema>
export type TestIndexerResponse = z.infer<typeof TestIndexerResponseSchema>
export type IndexerOptions = z.infer<typeof IndexerOptionsSchema>
export type UpdateIndexerOptionsInput = z.infer<typeof UpdateIndexerOptionsInputSchema>
