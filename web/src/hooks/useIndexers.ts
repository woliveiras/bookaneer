import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  type CreateIndexerInput,
  indexerApi,
  type UpdateIndexerInput,
  type UpdateIndexerOptionsInput,
} from "../lib/api"

export function useIndexers() {
  return useQuery({
    queryKey: ["indexers"],
    queryFn: () => indexerApi.list(),
    staleTime: 30 * 1000,
  })
}

export function useIndexer(id: number) {
  return useQuery({
    queryKey: ["indexer", id],
    queryFn: () => indexerApi.get(id),
    enabled: id > 0,
    staleTime: 30 * 1000,
  })
}

export function useCreateIndexer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateIndexerInput) => indexerApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["indexers"] })
    },
  })
}

export function useUpdateIndexer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateIndexerInput }) =>
      indexerApi.update(id, data),
    onSuccess: (indexer) => {
      queryClient.invalidateQueries({ queryKey: ["indexers"] })
      queryClient.setQueryData(["indexer", indexer.id], indexer)
    },
  })
}

export function useDeleteIndexer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => indexerApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["indexers"] })
    },
  })
}

export function useTestIndexer() {
  return useMutation({
    mutationFn: (data: CreateIndexerInput) => indexerApi.test(data),
  })
}

export interface SearchParams {
  q?: string
  author?: string
  title?: string
  isbn?: string
  category?: string
  limit?: number
  offset?: number
}

export function useSearch(params: SearchParams, enabled = true) {
  return useQuery({
    queryKey: ["search", params],
    queryFn: () => indexerApi.search(params),
    enabled: enabled && !!(params.q || params.author || params.title || params.isbn),
    staleTime: 60 * 1000, // 1 minute
    retry: 2,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 5000),
  })
}

export function useIndexerOptions() {
  return useQuery({
    queryKey: ["indexerOptions"],
    queryFn: () => indexerApi.getOptions(),
    staleTime: 60 * 1000,
  })
}

export function useUpdateIndexerOptions() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateIndexerOptionsInput) => indexerApi.updateOptions(data),
    onSuccess: (updated) => {
      queryClient.setQueryData(["indexerOptions"], updated)
    },
  })
}
