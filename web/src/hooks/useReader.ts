import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { readerApi, type SaveProgressInput } from "../lib/api"

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
