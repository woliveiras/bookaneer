export interface ReaderBookFile {
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
  bookTitle?: string
  authorName?: string
  coverUrl?: string
}

export interface ReadingProgress {
  id?: number
  bookFileId: number
  userId?: number
  position: string // EPUB CFI
  percentage: number
  updatedAt?: string
}

export interface SaveProgressInput {
  position: string
  percentage: number
}

export interface Bookmark {
  id: number
  bookFileId: number
  userId: number
  position: string
  title: string
  note: string
  createdAt: string
}

export interface CreateBookmarkInput {
  position: string
  title: string
  note?: string
}
