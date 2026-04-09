import type {
  ActiveCommand,
  BlocklistItem,
  HistoryEventType,
  HistoryItem,
  SearchCommandResponse,
  WantedResponse,
} from "../types"
import { fetchAPI } from "./client"

export const wantedApi = {
  getMissing: () => fetchAPI<WantedResponse>("/wanted/missing"),

  searchAllMissing: () =>
    fetchAPI<SearchCommandResponse>("/wanted/missing/search", {
      method: "POST",
    }),

  searchBook: (bookId: number) =>
    fetchAPI<SearchCommandResponse>(`/book/${bookId}/search`, {
      method: "POST",
    }),

  manualGrab: (data: {
    bookId: number
    downloadUrl: string
    releaseTitle?: string
    size?: number
    quality?: string
  }) =>
    fetchAPI<SearchCommandResponse>("/release", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getActiveCommands: () => fetchAPI<ActiveCommand[]>("/commands/active"),

  getRecentCommands: (limit?: number) => {
    const params = limit ? `?limit=${limit}` : ""
    return fetchAPI<ActiveCommand[]>(`/commands/recent${params}`)
  },
}

export const historyApi = {
  list: (params?: { limit?: number; eventType?: HistoryEventType }) => {
    const searchParams = new URLSearchParams()
    if (params?.limit) searchParams.set("limit", String(params.limit))
    if (params?.eventType) searchParams.set("eventType", params.eventType)
    const query = searchParams.toString()
    return fetchAPI<HistoryItem[]>(`/history${query ? `?${query}` : ""}`)
  },
}

export const blocklistApi = {
  list: () => fetchAPI<BlocklistItem[]>("/blocklist"),

  add: (data: { bookId: number; sourceTitle: string; quality?: string; reason?: string }) =>
    fetchAPI<void>("/blocklist", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  remove: (id: number) =>
    fetchAPI<void>(`/blocklist/${id}`, {
      method: "DELETE",
    }),
}
