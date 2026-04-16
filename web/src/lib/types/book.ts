export interface Book {
  id: number
  authorId: number
  title: string
  sortTitle: string
  foreignId: string
  isbn: string
  isbn13: string
  releaseDate: string
  overview: string
  imageUrl: string
  pageCount: number
  monitored: boolean
  userRating?: number // 1-5, undefined = unrated
  inWishlist: boolean
  addedAt: string
  updatedAt: string
  authorName?: string
  hasFile?: boolean
  fileFormat?: string
}

export interface Edition {
  id: number
  bookId: number
  foreignId: string
  title: string
  isbn: string
  isbn13: string
  format: string
  publisher: string
  releaseDate: string
  pageCount: number
  language: string
  monitored: boolean
}

export interface BookFile {
  id: number
  bookId: number
  editionId?: number
  path: string
  relativePath: string
  size: number
  format: string
  quality: string
  hash: string
  addedAt: string
  contentMismatch: boolean
}

export interface BookWithEditions extends Book {
  editions: Edition[]
  files?: BookFile[]
}

export interface CreateBookInput {
  authorId: number
  title: string
  sortTitle?: string
  foreignId?: string
  isbn?: string
  isbn13?: string
  releaseDate?: string
  overview?: string
  imageUrl?: string
  pageCount?: number
  monitored?: boolean
  userRating?: number
  inWishlist?: boolean
}

export interface ListBooksParams {
  authorId?: number
  monitored?: boolean
  missing?: boolean
  inWishlist?: boolean
  search?: string
  sortBy?: string
  sortDir?: string
  limit?: number
  offset?: number
}

export interface PaginatedResponse<T> {
  records: T[]
  totalRecords: number
}
