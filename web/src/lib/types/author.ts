export interface Author {
  id: number
  name: string
  sortName: string
  foreignId: string
  overview: string
  imageUrl: string
  status: string
  monitored: boolean
  path: string
  addedAt: string
  updatedAt: string
  bookCount?: number
  bookFileCount?: number
}

export interface AuthorStats {
  bookCount: number
  bookFileCount: number
  missingBooks: number
  totalSizeBytes: number
}

export interface CreateAuthorInput {
  name: string
  sortName?: string
  foreignId?: string
  overview?: string
  imageUrl?: string
  status?: string
  monitored?: boolean
  path?: string
}

export interface UpdateAuthorInput {
  name?: string
  sortName?: string
  foreignId?: string
  overview?: string
  imageUrl?: string
  status?: string
  monitored?: boolean
  path?: string
}

export interface ListAuthorsParams {
  monitored?: boolean
  status?: string
  search?: string
  sortBy?: string
  sortDir?: string
  limit?: number
  offset?: number
}
