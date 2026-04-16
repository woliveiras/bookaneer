import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useRef } from "react"
import type { HistoryEventType } from "../lib/api"
import { blocklistApi, historyApi, queueApi, wantedApi } from "../lib/api"

export function useSearchBook() {
  return useMutation({
    mutationFn: wantedApi.searchBook,
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

export function useIndexerGrab() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: wantedApi.indexerGrab,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["queue"] })
      queryClient.invalidateQueries({ queryKey: ["wanted"] })
    },
  })
}

// Download queue hooks
export function useDownloadQueue() {
  const queryClient = useQueryClient()
  const prevCompletedRef = useRef<Set<number>>(new Set())

  return useQuery({
    queryKey: ["queue"],
    queryFn: async () => {
      const data = await queueApi.list()
      const currentCompleted = new Set(
        data.filter((item) => item.status === "completed").map((item) => item.id),
      )
      // If there are newly completed items, invalidate related caches
      const hasNew = [...currentCompleted].some((id) => !prevCompletedRef.current.has(id))
      if (hasNew) {
        queryClient.invalidateQueries({ queryKey: ["wanted"] })
        queryClient.invalidateQueries({ queryKey: ["books"] })
      }
      prevCompletedRef.current = currentCompleted
      return data
    },
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

export function useRecentCommands(limit = 10) {
  return useQuery({
    queryKey: ["commands", "recent", limit],
    queryFn: () => wantedApi.getRecentCommands(limit),
    refetchInterval: 5000, // Refresh every 5 seconds
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

export function useRetryDownload() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: queueApi.retry,
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

export function useReportWrongContent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: wantedApi.reportWrongContent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["books"] })
      queryClient.invalidateQueries({ queryKey: ["book"] })
      queryClient.invalidateQueries({ queryKey: ["queue"] })
      queryClient.invalidateQueries({ queryKey: ["blocklist"] })
      queryClient.invalidateQueries({ queryKey: ["history"] })
    },
  })
}
