import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  type CreateDownloadClientInput,
  downloadClientApi,
  queueApi,
} from "../lib/api"

// Download Clients hooks
export function useDownloadClients() {
  return useQuery({
    queryKey: ["downloadClients"],
    queryFn: downloadClientApi.list,
  })
}

export function useDownloadClient(id: number) {
  return useQuery({
    queryKey: ["downloadClient", id],
    queryFn: () => downloadClientApi.get(id),
    enabled: id > 0,
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

// Queue hooks
export function useQueue() {
  return useQuery({
    queryKey: ["queue"],
    queryFn: queueApi.list,
    refetchInterval: 5000, // Refresh every 5 seconds
  })
}
