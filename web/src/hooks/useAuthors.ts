import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  authorApi,
  type CreateAuthorInput,
  type ListAuthorsParams,
  type UpdateAuthorInput,
} from "../lib/api"

export function useAuthors(params?: ListAuthorsParams) {
  return useQuery({
    queryKey: ["authors", params],
    queryFn: () => authorApi.list(params),
    staleTime: 30 * 1000, // 30 seconds
  })
}

export function useAuthor(id: number) {
  return useQuery({
    queryKey: ["author", id],
    queryFn: () => authorApi.get(id),
    enabled: id > 0,
    staleTime: 30 * 1000,
  })
}

export function useAuthorStats(id: number) {
  return useQuery({
    queryKey: ["author", id, "stats"],
    queryFn: () => authorApi.getStats(id),
    enabled: id > 0,
    staleTime: 30 * 1000,
  })
}

export function useCreateAuthor() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateAuthorInput) => authorApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["authors"] })
    },
  })
}

export function useUpdateAuthor() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateAuthorInput }) =>
      authorApi.update(id, data),
    onSuccess: (author) => {
      queryClient.invalidateQueries({ queryKey: ["authors"] })
      queryClient.setQueryData(["author", author.id], author)
    },
  })
}

export function useDeleteAuthor() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, deleteFiles }: { id: number; deleteFiles?: boolean }) =>
      authorApi.delete(id, deleteFiles),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["authors"] })
      queryClient.removeQueries({ queryKey: ["author", id] })
    },
  })
}
