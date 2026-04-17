import type {
  CreateIndexerInput,
  Indexer,
  IndexerOptions,
  SearchResponse,
  TestIndexerResponse,
  UpdateIndexerInput,
  UpdateIndexerOptionsInput,
} from "../schemas"
import { API_BASE, fetchAPI, getStoredApiKey } from "./client"

export const indexerApi = {
  list: () => fetchAPI<Indexer[]>("/indexer"),

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
