export interface MetadataAuthorResult {
  foreignId: string
  name: string
  birthYear?: number
  deathYear?: number
  photoUrl?: string
  worksCount?: number
  provider: string
}

export interface MetadataBookResult {
  foreignId: string
  title: string
  authors?: string[]
  publishedYear?: number
  coverUrl?: string
  isbn10?: string
  isbn13?: string
  provider: string
}

export interface MetadataAuthor {
  foreignId: string
  name: string
  sortName?: string
  bio?: string
  birthDate?: string
  deathDate?: string
  photoUrl?: string
  website?: string
  wikipedia?: string
  nationality?: string
  provider: string
  links?: { type: string; url: string }[]
}

export interface MetadataBook {
  foreignId: string
  title: string
  subtitle?: string
  authors?: string[]
  authorIds?: string[]
  description?: string
  publishedDate?: string
  publisher?: string
  pageCount?: number
  language?: string
  isbn10?: string
  isbn13?: string
  asin?: string
  coverUrl?: string
  genres?: string[]
  subjects?: string[]
  series?: string
  seriesPosition?: number
  averageRating?: number
  ratingsCount?: number
  provider: string
  links?: { type: string; url: string }[]
}

export interface MetadataSearchResponse<T> {
  results: T[]
  total: number
}

export interface DigitalLibraryResult {
  id: string
  title: string
  authors?: string[]
  publisher?: string
  year?: number
  language?: string
  format: string
  size: number
  isbn?: string
  coverUrl?: string
  downloadUrl?: string
  infoUrl?: string
  provider: string
  score?: number
}

export interface DigitalLibrarySearchResponse {
  results: DigitalLibraryResult[]
  total: number
}
