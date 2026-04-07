const API_BASE = "/api/v1"

// Generic fetch wrapper
async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: res.statusText }))
    throw new Error(error.message || res.statusText)
  }

  return res.json()
}

// Author types
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
  path: string
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

// Book types
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
  addedAt: string
  updatedAt: string
  authorName?: string
  hasFile?: boolean
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
}

export interface ListBooksParams {
  authorId?: number
  monitored?: boolean
  missing?: boolean
  search?: string
  sortBy?: string
  sortDir?: string
  limit?: number
  offset?: number
}

// Paginated response
export interface PaginatedResponse<T> {
  records: T[]
  totalRecords: number
}

// Author API
export const authorApi = {
  list: (params?: ListAuthorsParams) => {
    const searchParams = new URLSearchParams()
    if (params?.monitored !== undefined) searchParams.set("monitored", String(params.monitored))
    if (params?.status) searchParams.set("status", params.status)
    if (params?.search) searchParams.set("search", params.search)
    if (params?.sortBy) searchParams.set("sortBy", params.sortBy)
    if (params?.sortDir) searchParams.set("sortDir", params.sortDir)
    if (params?.limit) searchParams.set("limit", String(params.limit))
    if (params?.offset) searchParams.set("offset", String(params.offset))
    const query = searchParams.toString()
    return fetchAPI<PaginatedResponse<Author>>(`/author${query ? `?${query}` : ""}`)
  },

  get: (id: number) => fetchAPI<Author>(`/author/${id}`),

  getStats: (id: number) => fetchAPI<AuthorStats>(`/author/${id}/stats`),

  create: (data: CreateAuthorInput) =>
    fetchAPI<Author>("/author", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: UpdateAuthorInput) =>
    fetchAPI<Author>(`/author/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetchAPI<void>(`/author/${id}`, {
      method: "DELETE",
    }),
}

// Book API
export const bookApi = {
  list: (params?: ListBooksParams) => {
    const searchParams = new URLSearchParams()
    if (params?.authorId) searchParams.set("authorId", String(params.authorId))
    if (params?.monitored !== undefined) searchParams.set("monitored", String(params.monitored))
    if (params?.missing) searchParams.set("missing", "true")
    if (params?.search) searchParams.set("search", params.search)
    if (params?.sortBy) searchParams.set("sortBy", params.sortBy)
    if (params?.sortDir) searchParams.set("sortDir", params.sortDir)
    if (params?.limit) searchParams.set("limit", String(params.limit))
    if (params?.offset) searchParams.set("offset", String(params.offset))
    const query = searchParams.toString()
    return fetchAPI<PaginatedResponse<Book>>(`/book${query ? `?${query}` : ""}`)
  },

  get: (id: number) => fetchAPI<BookWithEditions>(`/book/${id}`),

  create: (data: CreateBookInput) =>
    fetchAPI<Book>("/book", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<CreateBookInput>) =>
    fetchAPI<Book>(`/book/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetchAPI<void>(`/book/${id}`, {
      method: "DELETE",
    }),
}
