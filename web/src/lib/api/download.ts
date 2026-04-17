import type {
  CreateDownloadClientInput,
  DownloadClient,
  QueueItem,
  TestDownloadClientResponse,
} from "../schemas"
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

  remove: (id: number) =>
    fetch(`${API_BASE}/queue/${id}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to remove from queue")
    }),

  retry: (id: number) =>
    fetchAPI<void>(`/queue/${id}/retry`, {
      method: "POST",
    }),
}
