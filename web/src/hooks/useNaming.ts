import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { namingApi } from "../lib/api"
import type { NamingPreviewInput, NamingSettingsInput } from "../lib/types"

export function useNamingSettings() {
  return useQuery({
    queryKey: ["naming", "settings"],
    queryFn: () => namingApi.getSettings(),
  })
}

export function useUpdateNamingSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: NamingSettingsInput) => namingApi.updateSettings(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["naming"] })
    },
  })
}

export function useNamingPreview(data: NamingPreviewInput | null) {
  return useQuery({
    queryKey: ["naming", "preview", data],
    queryFn: () => namingApi.preview(data!),
    enabled: !!data,
  })
}

export function usePreviewRenameAll() {
  return useMutation({
    mutationFn: () => namingApi.previewRenameAll(),
  })
}

export function useRenameAll() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => namingApi.renameAll(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["books"] })
    },
  })
}
