import type {
  DigitalLibrarySearchResponse,
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

  getBook: (foreignId: string, provider?: string) => {
    const params = provider ? `?provider=${encodeURIComponent(provider)}` : ""
    return fetchAPI<MetadataBook>(`/metadata/books/${encodeURIComponent(foreignId)}${params}`)
  },
}

export const digitalLibraryApi = {
  search: (query: string) =>
    fetchAPI<DigitalLibrarySearchResponse>(`/digitallibrary/search?q=${encodeURIComponent(query)}`),
}
