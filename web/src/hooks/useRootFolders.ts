import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { type CreateRootFolderInput, rootFolderApi, type UpdateRootFolderInput } from "../lib/api"

export function useRootFolders() {
  return useQuery({
    queryKey: ["rootFolders"],
    queryFn: rootFolderApi.list,
  })
}

export function useRootFolder(id: number) {
  return useQuery({
    queryKey: ["rootFolder", id],
    queryFn: () => rootFolderApi.get(id),
    enabled: id > 0,
  })
}

export function useCreateRootFolder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: rootFolderApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["rootFolders"] })
    },
  })
}

export function useUpdateRootFolder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateRootFolderInput }) =>
      rootFolderApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["rootFolders"] })
    },
  })
}

export function useMigrateRootFolder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, newPath }: { id: number; newPath: string }) =>
      rootFolderApi.migrate(id, newPath),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["rootFolders"] })
      queryClient.invalidateQueries({ queryKey: ["authors"] })
    },
  })
}

export function useDeleteRootFolder() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: rootFolderApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["rootFolders"] })
    },
  })
}
