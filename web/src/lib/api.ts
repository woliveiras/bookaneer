const API_BASE = "/api/v1"
const API_KEY_STORAGE_KEY = "bookaneer_api_key"

// Get stored API key
export function getStoredApiKey(): string | null {
  return localStorage.getItem(API_KEY_STORAGE_KEY)
}

// Set stored API key
export function setStoredApiKey(apiKey: string): void {
  localStorage.setItem(API_KEY_STORAGE_KEY, apiKey)
}

// Clear stored API key
export function clearStoredApiKey(): void {
  localStorage.removeItem(API_KEY_STORAGE_KEY)
}

// Generic fetch wrapper with auth
async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const apiKey = getStoredApiKey()
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options?.headers,
  }
  if (apiKey) {
    ;(headers as Record<string, string>)["X-Api-Key"] = apiKey
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: res.statusText }))
    throw new Error(error.message || res.statusText)
  }

  return res.json()
}

// Auth types
export interface User {
  id: number
  username: string
  role: string
  apiKey?: string
  createdAt: string
}

export interface LoginResponse {
  user: User
  apiKey: string
}

// Auth API
export const authApi = {
  login: (apiKey: string) =>
    fetchAPI<User>("/auth/me", {
      headers: { "X-Api-Key": apiKey },
    }),

  loginWithCredentials: (username: string, password: string) =>
    fetchAPI<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    }),

  me: () => fetchAPI<User>("/auth/me"),

  logout: () =>
    fetchAPI<{ status: string }>("/auth/logout", {
      method: "POST",
    }),
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

// Metadata types (external provider results)
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

// Metadata API
export const metadataApi = {
  searchAuthors: (query: string) =>
    fetchAPI<MetadataSearchResponse<MetadataAuthorResult>>(`/metadata/authors?q=${encodeURIComponent(query)}`),

  searchBooks: (query: string) =>
    fetchAPI<MetadataSearchResponse<MetadataBookResult>>(`/metadata/books?q=${encodeURIComponent(query)}`),

  getAuthor: (foreignId: string, provider?: string) => {
    const params = provider ? `?provider=${encodeURIComponent(provider)}` : ""
    return fetchAPI<MetadataAuthor>(`/metadata/authors/${encodeURIComponent(foreignId)}${params}`)
  },

  getBook: (foreignId: string, provider?: string) => {
    const params = provider ? `?provider=${encodeURIComponent(provider)}` : ""
    return fetchAPI<MetadataBook>(`/metadata/books/${encodeURIComponent(foreignId)}${params}`)
  },

  lookupISBN: (isbn: string) =>
    fetchAPI<MetadataBook>(`/metadata/isbn/${encodeURIComponent(isbn)}`),

  getProviders: () =>
    fetchAPI<{ providers: string[] }>("/metadata/providers"),
}

// Settings types
export interface GeneralSettings {
  apiKey: string
  bindAddress: string
  port: number
  dataDir: string
  libraryDir: string
  logLevel: string
}

// Settings API
export const settingsApi = {
  getGeneral: () => fetchAPI<GeneralSettings>("/settings/general"),
}

// Reader types
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

// Bookmark types
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

// Reader API
export const readerApi = {
  getBookFile: (id: number) => fetchAPI<ReaderBookFile>(`/reader/${id}`),

  getContentUrl: (id: number) => {
    const apiKey = getStoredApiKey()
    const base = `${API_BASE}/reader/${id}/content`
    return apiKey ? `${base}?key=${encodeURIComponent(apiKey)}` : base
  },

  getProgress: (id: number) => fetchAPI<ReadingProgress>(`/reader/${id}/progress`),

  saveProgress: (id: number, data: SaveProgressInput) =>
    fetchAPI<ReadingProgress>(`/reader/${id}/progress`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  listBookmarks: (id: number) => fetchAPI<Bookmark[]>(`/reader/${id}/bookmarks`),

  createBookmark: (id: number, data: CreateBookmarkInput) =>
    fetchAPI<Bookmark>(`/reader/${id}/bookmarks`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  deleteBookmark: (bookFileId: number, bookmarkId: number) =>
    fetch(`${API_BASE}/reader/${bookFileId}/bookmarks/${bookmarkId}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to delete bookmark")
    }),
}

// Indexer types
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

export interface SearchResponse {
  results: SearchResult[]
  total: number
}

export interface TestIndexerResponse {
  success: boolean
  message: string
}

// Indexer Options types
export interface IndexerOptions {
  minimumAge: number         // Minutes
  retention: number          // Days (0 = unlimited)
  maximumSize: number        // MB (0 = unlimited)
  rssSyncInterval: number    // Minutes (0 = disabled)
  preferIndexerFlags: boolean
  availabilityDelay: number  // Days
  updatedAt: string
}

export type UpdateIndexerOptionsInput = Omit<IndexerOptions, "updatedAt">

// Indexer API
export const indexerApi = {
  list: () => fetchAPI<Indexer[]>("/indexer"),

  get: (id: number) => fetchAPI<Indexer>(`/indexer/${id}`),

  create: (data: CreateIndexerInput) =>
    fetchAPI<Indexer>("/indexer", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: UpdateIndexerInput) =>
    fetchAPI<Indexer>(`/indexer/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetch(`${API_BASE}/indexer/${id}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to delete indexer")
    }),

  test: (data: CreateIndexerInput) =>
    fetchAPI<TestIndexerResponse>("/indexer/test", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getOptions: () => fetchAPI<IndexerOptions>("/indexer/options"),

  updateOptions: (data: UpdateIndexerOptionsInput) =>
    fetchAPI<IndexerOptions>("/indexer/options", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  search: (params: {
    q?: string
    author?: string
    title?: string
    isbn?: string
    category?: string
    limit?: number
    offset?: number
  }) => {
    const searchParams = new URLSearchParams()
    if (params.q) searchParams.set("q", params.q)
    if (params.author) searchParams.set("author", params.author)
    if (params.title) searchParams.set("title", params.title)
    if (params.isbn) searchParams.set("isbn", params.isbn)
    if (params.category) searchParams.set("category", params.category)
    if (params.limit) searchParams.set("limit", params.limit.toString())
    if (params.offset) searchParams.set("offset", params.offset.toString())
    return fetchAPI<SearchResponse>(`/search?${searchParams.toString()}`)
  },
}

// Download client types
export type DownloadClientType = "sabnzbd" | "qbittorrent" | "transmission" | "blackhole"

export interface DownloadClient {
  id: number
  name: string
  type: DownloadClientType
  host: string
  port: number
  useTls: boolean
  username: string
  password: string
  apiKey: string
  category: string
  recentPriority: number
  olderPriority: number
  removeCompletedAfter: number
  enabled: boolean
  priority: number
  nzbFolder: string
  torrentFolder: string
  watchFolder: string
  createdAt: string
  updatedAt: string
}

export interface CreateDownloadClientInput {
  name: string
  type: DownloadClientType
  host?: string
  port?: number
  useTls?: boolean
  username?: string
  password?: string
  apiKey?: string
  category?: string
  recentPriority?: number
  olderPriority?: number
  removeCompletedAfter?: number
  enabled?: boolean
  priority?: number
  nzbFolder?: string
  torrentFolder?: string
  watchFolder?: string
}

export type DownloadStatus = "queued" | "downloading" | "paused" | "completed" | "seeding" | "failed" | "extracted" | "processing"

export interface QueueItem {
  id: string
  name: string
  status: DownloadStatus
  progress: number
  size: number
  downloadedSize: number
  speed: number
  eta: number
  seeders?: number
  leechers?: number
  ratio?: number
  savePath?: string
  category?: string
  errorMessage?: string
  addedAt: string
  completedAt?: string
  clientId: number
}

export type GrabStatus = "pending" | "sent" | "downloading" | "completed" | "failed" | "imported"

export interface Grab {
  id: number
  bookId: number
  indexerId: number
  releaseTitle: string
  downloadUrl: string
  size: number
  quality: string
  clientId: number
  downloadId: string
  status: GrabStatus
  errorMessage: string
  grabbedAt: string
  completedAt?: string
}

export interface CreateGrabInput {
  bookId: number
  indexerId: number
  releaseTitle: string
  downloadUrl: string
  size?: number
  quality?: string
  clientId: number
}

export interface TestDownloadClientResponse {
  success: boolean
  message: string
}

// Download client API
export const downloadClientApi = {
  list: () => fetchAPI<DownloadClient[]>("/downloadclient"),

  get: (id: number) => fetchAPI<DownloadClient>(`/downloadclient/${id}`),

  create: (data: CreateDownloadClientInput) =>
    fetchAPI<DownloadClient>("/downloadclient", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: CreateDownloadClientInput) =>
    fetchAPI<DownloadClient>(`/downloadclient/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetch(`${API_BASE}/downloadclient/${id}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to delete download client")
    }),

  test: (data: CreateDownloadClientInput) =>
    fetchAPI<TestDownloadClientResponse>("/downloadclient/test", {
      method: "POST",
      body: JSON.stringify(data),
    }),
}

// Queue API
export const queueApi = {
  list: () => fetchAPI<QueueItem[]>("/queue"),

  listByClient: (clientId: number) => fetchAPI<QueueItem[]>(`/queue/${clientId}`),
}

// Grab API
export const grabApi = {
  list: (params?: { bookId?: number; status?: GrabStatus; limit?: number }) => {
    const searchParams = new URLSearchParams()
    if (params?.bookId) searchParams.set("bookId", params.bookId.toString())
    if (params?.status) searchParams.set("status", params.status)
    if (params?.limit) searchParams.set("limit", params.limit.toString())
    const query = searchParams.toString()
    return fetchAPI<Grab[]>(`/grab${query ? `?${query}` : ""}`)
  },

  create: (data: CreateGrabInput) =>
    fetchAPI<Grab>("/grab", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  send: (id: number) =>
    fetchAPI<Grab>(`/grab/${id}/send`, {
      method: "POST",
    }),
}
