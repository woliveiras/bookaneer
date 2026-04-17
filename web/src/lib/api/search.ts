import type {
  DigitalLibrarySearchResponse,
  MetadataAuthor,
  MetadataAuthorResult,
  MetadataBook,
  MetadataBookResult,
  MetadataSearchResponse,
} from "../schemas"
import { fetchAPI } from "./client"

export const metadataApi = {
  searchAuthors: (query: string) =>
    fetchAPI<MetadataSearchResponse<MetadataAuthorResult>>(
      `/metadata/authors?q=${encodeURIComponent(query)}`,
    ),

  searchBooks: (query: string) =>
    fetchAPI<MetadataSearchResponse<MetadataBookResult>>(
      `/metadata/books?q=${encodeURIComponent(query)}`,
    ),

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

  getProviders: () => fetchAPI<{ providers: string[] }>("/metadata/providers"),
}

export const digitalLibraryApi = {
  search: (query: string) =>
    fetchAPI<DigitalLibrarySearchResponse>(`/digitallibrary/search?q=${encodeURIComponent(query)}`),

  getProviders: () => fetchAPI<{ providers: string[] }>("/digitallibrary/providers"),
}
