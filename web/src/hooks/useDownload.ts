import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  type CreateDownloadClientInput,
  downloadClientApi,
} from "../lib/api"

// Download Clients hooks
export function useDownloadClients() {
  return useQuery({
    queryKey: ["downloadClients"],
    queryFn: downloadClientApi.list,
  })
}

export function useCreateDownloadClient() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: downloadClientApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["downloadClients"] })
    },
  })
}

export function useUpdateDownloadClient() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: CreateDownloadClientInput }) =>
      downloadClientApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["downloadClients"] })
    },
  })
}

export function useDeleteDownloadClient() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: downloadClientApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["downloadClients"] })
    },
  })
}

export function useTestDownloadClient() {
  return useMutation({
    mutationFn: downloadClientApi.test,
  })
}

