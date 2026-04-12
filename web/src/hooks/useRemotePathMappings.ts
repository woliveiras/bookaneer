import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  remotePathMappingApi,
  type CreateRemotePathMappingInput,
  type UpdateRemotePathMappingInput,
} from "../lib/api"

export function useRemotePathMappings() {
  return useQuery({
    queryKey: ["remotePathMappings"],
    queryFn: remotePathMappingApi.list,
    staleTime: 30 * 1000,
  })
}

export function useCreateRemotePathMapping() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateRemotePathMappingInput) => remotePathMappingApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["remotePathMappings"] })
    },
  })
}

export function useUpdateRemotePathMapping() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateRemotePathMappingInput }) =>
      remotePathMappingApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["remotePathMappings"] })
    },
  })
}

export function useDeleteRemotePathMapping() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => remotePathMappingApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["remotePathMappings"] })
    },
  })
}
