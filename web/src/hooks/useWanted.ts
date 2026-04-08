import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { wantedApi, queueApi, historyApi, blocklistApi } from "../lib/api"
import type { HistoryEventType } from "../lib/api"

// Missing books hooks
export function useWantedMissing() {
  return useQuery({
    queryKey: ["wanted", "missing"],
    queryFn: wantedApi.getMissing,
  })
}

export function useSearchAllMissing() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: wantedApi.searchAllMissing,
    onSuccess: () => {
      // Invalidate queue after starting search
      queryClient.invalidateQueries({ queryKey: ["queue"] })
    },
  })
}

export function useSearchBook() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: wantedApi.searchBook,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["queue"] })
      queryClient.invalidateQueries({ queryKey: ["wanted"] })
    },
  })
}

export function useManualGrab() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: wantedApi.manualGrab,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["queue"] })
      queryClient.invalidateQueries({ queryKey: ["wanted"] })
    },
  })
}

// Download queue hooks
export function useDownloadQueue() {
  return useQuery({
    queryKey: ["queue"],
    queryFn: queueApi.list,
    refetchInterval: 5000, // Refresh every 5 seconds
  })
}

export function useActiveCommands() {
  return useQuery({
    queryKey: ["commands", "active"],
    queryFn: wantedApi.getActiveCommands,
    refetchInterval: 2000, // Refresh every 2 seconds for active commands
  })
}

export function useRemoveFromQueue() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: queueApi.remove,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["queue"] })
    },
  })
}

// History hooks
export function useHistory(params?: { limit?: number; eventType?: HistoryEventType }) {
  return useQuery({
    queryKey: ["history", params],
    queryFn: () => historyApi.list(params),
  })
}

// Blocklist hooks
export function useBlocklist() {
  return useQuery({
    queryKey: ["blocklist"],
    queryFn: blocklistApi.list,
  })
}

export function useAddToBlocklist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: blocklistApi.add,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["blocklist"] })
    },
  })
}

export function useRemoveFromBlocklist() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: blocklistApi.remove,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["blocklist"] })
    },
  })
}
