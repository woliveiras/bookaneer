import { fetchAPI, getStoredApiKey, API_BASE } from "./client"
import type { ReaderBookFile, ReadingProgress, SaveProgressInput, Bookmark, CreateBookmarkInput } from "../types"

export const readerApi = {
  getBookFile: (id: number) => fetchAPI<ReaderBookFile>(`/reader/${id}`),

  getContentUrl: (id: number) => {
    const apiKey = getStoredApiKey()
    const base = `${API_BASE}/reader/${id}/content`
    return apiKey ? `${base}?key=${encodeURIComponent(apiKey)}` : base
  },

  getProgress: (id: number) => fetchAPI<ReadingProgress>(`/reader/${id}/progress`),

  saveProgress: (id: number, data: SaveProgressInput) =>
    fetchAPI<ReadingProgress>(`/reader/${id}/progress`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  listBookmarks: (id: number) => fetchAPI<Bookmark[]>(`/reader/${id}/bookmarks`),

  createBookmark: (id: number, data: CreateBookmarkInput) =>
    fetchAPI<Bookmark>(`/reader/${id}/bookmarks`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  deleteBookmark: (bookFileId: number, bookmarkId: number) =>
    fetch(`${API_BASE}/reader/${bookFileId}/bookmarks/${bookmarkId}`, {
      method: "DELETE",
      headers: {
        "X-Api-Key": getStoredApiKey() || "",
      },
    }).then((res) => {
      if (!res.ok) throw new Error("Failed to delete bookmark")
    }),
}
