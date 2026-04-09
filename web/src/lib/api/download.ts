import type {
  CreateDownloadClientInput,
  CreateGrabInput,
  DownloadClient,
  Grab,
  GrabStatus,
  QueueItem,
  TestDownloadClientResponse,
} from "../types"
import { API_BASE, fetchAPI, getStoredApiKey } from "./client"

export const downloadClientApi = {
  list: () => fetchAPI<DownloadClient[]>("/downloadclient"),

  get: (id: number) => fetchAPI<DownloadClient>(`/downloadclient/${id}`),

  create: (data: CreateDownloadClientInput) =>
    fetchAPI<DownloadClient>("/downloadclient", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: CreateDownloadClientInput) =>
    fetchAPI<DownloadClient>(`/downloadclient/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetch(`${API_BASE}/downloadclient/${id}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to delete download client")
    }),

  test: (data: CreateDownloadClientInput) =>
    fetchAPI<TestDownloadClientResponse>("/downloadclient/test", {
      method: "POST",
      body: JSON.stringify(data),
    }),
}

export const queueApi = {
  list: () => fetchAPI<QueueItem[]>("/queue"),

  listByClient: (clientId: number) => fetchAPI<QueueItem[]>(`/queue/${clientId}`),

  remove: (id: number) =>
    fetch(`${API_BASE}/queue/${id}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to remove from queue")
    }),
}

export const grabApi = {
  list: (params?: { bookId?: number; status?: GrabStatus; limit?: number }) => {
    const searchParams = new URLSearchParams()
    if (params?.bookId) searchParams.set("bookId", params.bookId.toString())
    if (params?.status) searchParams.set("status", params.status)
    if (params?.limit) searchParams.set("limit", params.limit.toString())
    const query = searchParams.toString()
    return fetchAPI<Grab[]>(`/grab${query ? `?${query}` : ""}`)
  },

  create: (data: CreateGrabInput) =>
    fetchAPI<Grab>("/grab", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  send: (id: number) =>
    fetchAPI<Grab>(`/grab/${id}/send`, {
      method: "POST",
    }),
}
