export interface Indexer {
  id: number
  name: string
  type: "newznab" | "torznab"
  baseUrl: string
  apiPath: string
  apiKey: string
  categories: string
  priority: number
  enabled: boolean
  enableRss: boolean
  enableAutomaticSearch: boolean
  enableInteractiveSearch: boolean
  additionalParameters: string
  minimumSeeders: number
  seedRatio?: number | null
  seedTime?: number | null
  createdAt: string
  updatedAt: string
}

export interface CreateIndexerInput {
  name: string
  type: "newznab" | "torznab"
  baseUrl: string
  apiPath?: string
  apiKey: string
  categories?: string
  priority?: number
  enabled?: boolean
  enableRss?: boolean
  enableAutomaticSearch?: boolean
  enableInteractiveSearch?: boolean
  additionalParameters?: string
  minimumSeeders?: number
  seedRatio?: number | null
  seedTime?: number | null
}

export interface UpdateIndexerInput extends CreateIndexerInput {
  id?: number
}

export interface SearchResult {
  guid: string
  title: string
  description?: string
  size: number
  pubDate: string
  category?: string
  categoryId?: string
  downloadUrl: string
  infoUrl?: string
  comments?: number
  seeders?: number
  leechers?: number
  grabs?: number
  quality?: string
  qualityRank?: number
  indexerId: number
  indexerName: string
}

import type { ColumnConfig } from "./search"

export interface SearchResponse {
  results: SearchResult[]
  total: number
  columnConfig?: ColumnConfig
}

export interface TestIndexerResponse {
  success: boolean
  message: string
}

export interface IndexerOptions {
  minimumAge: number // Minutes
  retention: number // Days (0 = unlimited)
  maximumSize: number // MB (0 = unlimited)
  rssSyncInterval: number // Minutes (0 = disabled)
  preferIndexerFlags: boolean
  availabilityDelay: number // Days
  updatedAt: string
}

export type UpdateIndexerOptionsInput = Omit<IndexerOptions, "updatedAt">
