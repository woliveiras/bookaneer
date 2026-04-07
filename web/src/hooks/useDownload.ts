import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { downloadClientApi, queueApi, grabApi, type CreateDownloadClientInput, type CreateGrabInput, type GrabStatus } from "../lib/api"

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

export function useClientQueue(clientId: number) {
  return useQuery({
    queryKey: ["queue", clientId],
    queryFn: () => queueApi.listByClient(clientId),
    enabled: clientId > 0,
    refetchInterval: 5000,
  })
}

// Grab hooks
export function useGrabs(params?: { bookId?: number; status?: GrabStatus; limit?: number }) {
  return useQuery({
    queryKey: ["grabs", params],
    queryFn: () => grabApi.list(params),
  })
}

export function useCreateGrab() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: grabApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["grabs"] })
      queryClient.invalidateQueries({ queryKey: ["queue"] })
    },
  })
}

export function useSendGrab() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: grabApi.send,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["grabs"] })
      queryClient.invalidateQueries({ queryKey: ["queue"] })
    },
  })
}
