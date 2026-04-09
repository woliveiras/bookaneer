import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { type CreateBookmarkInput, readerApi, type SaveProgressInput } from "../lib/api"

export function useReaderBookFile(id: number) {
  return useQuery({
    queryKey: ["reader", "file", id],
    queryFn: () => readerApi.getBookFile(id),
    enabled: id > 0,
    staleTime: 60 * 1000,
  })
}

export function useReadingProgress(bookFileId: number) {
  return useQuery({
    queryKey: ["reader", "progress", bookFileId],
    queryFn: () => readerApi.getProgress(bookFileId),
    enabled: bookFileId > 0,
    staleTime: 10 * 1000,
  })
}

export function useSaveProgress(bookFileId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: SaveProgressInput) => readerApi.saveProgress(bookFileId, data),
    onSuccess: (progress) => {
      queryClient.setQueryData(["reader", "progress", bookFileId], progress)
    },
  })
}

// Bookmark hooks

export function useBookmarks(bookFileId: number) {
  return useQuery({
    queryKey: ["reader", "bookmarks", bookFileId],
    queryFn: () => readerApi.listBookmarks(bookFileId),
    enabled: bookFileId > 0,
    staleTime: 30 * 1000,
  })
}

export function useCreateBookmark(bookFileId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateBookmarkInput) => readerApi.createBookmark(bookFileId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reader", "bookmarks", bookFileId] })
    },
  })
}

export function useDeleteBookmark(bookFileId: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (bookmarkId: number) => readerApi.deleteBookmark(bookFileId, bookmarkId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["reader", "bookmarks", bookFileId] })
    },
  })
}
