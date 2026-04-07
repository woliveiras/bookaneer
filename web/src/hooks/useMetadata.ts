import { useQuery } from "@tanstack/react-query"
import { metadataApi } from "../lib/api"

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

// Get author details from metadata provider
export function useMetadataAuthor(foreignId: string, provider?: string, enabled = true) {
  return useQuery({
    queryKey: ["metadata", "author", foreignId, provider],
    queryFn: () => metadataApi.getAuthor(foreignId, provider),
    enabled: enabled && !!foreignId,
    staleTime: 30 * 60 * 1000, // 30 minutes
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

// Lookup book by ISBN
export function useMetadataISBN(isbn: string, enabled = true) {
  return useQuery({
    queryKey: ["metadata", "isbn", isbn],
    queryFn: () => metadataApi.lookupISBN(isbn),
    enabled: enabled && isbn.length >= 10,
    staleTime: 30 * 60 * 1000, // 30 minutes
  })
}

// Get available metadata providers
export function useMetadataProviders() {
  return useQuery({
    queryKey: ["metadata", "providers"],
    queryFn: () => metadataApi.getProviders(),
    staleTime: 60 * 60 * 1000, // 1 hour
  })
}
