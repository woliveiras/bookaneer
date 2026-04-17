import { useQuery } from "@tanstack/react-query"
import { digitalLibraryApi, metadataApi } from "../lib/api"

// Search authors across metadata providers
export function useMetadataSearchAuthors(query: string, enabled = true) {
  return useQuery({
    queryKey: ["metadata", "authors", "search", query],
    queryFn: () => metadataApi.searchAuthors(query),
    enabled: enabled && query.length >= 2,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

// Search books across metadata providers
export function useMetadataSearchBooks(query: string, enabled = true) {
  return useQuery({
    queryKey: ["metadata", "books", "search", query],
    queryFn: () => metadataApi.searchBooks(query),
    enabled: enabled && query.length >= 2,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

// Get book details from metadata provider
export function useMetadataBook(foreignId: string, provider?: string, enabled = true) {
  return useQuery({
    queryKey: ["metadata", "book", foreignId, provider],
    queryFn: () => metadataApi.getBook(foreignId, provider),
    enabled: enabled && !!foreignId,
    staleTime: 30 * 60 * 1000, // 30 minutes
  })
}

// Search digital libraries (Anna's Archive, LibGen, Internet Archive)
export function useDigitalLibrarySearch(query: string, enabled = true) {
  return useQuery({
    queryKey: ["digitallibrary", "search", query],
    queryFn: () => digitalLibraryApi.search(query),
    enabled: enabled && query.length >= 2,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 5000),
  })
}

