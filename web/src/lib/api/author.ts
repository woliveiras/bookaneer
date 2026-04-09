import { fetchAPI } from "./client"
import type { Author, AuthorStats, CreateAuthorInput, UpdateAuthorInput, ListAuthorsParams, PaginatedResponse } from "../types"

export const authorApi = {
  list: (params?: ListAuthorsParams) => {
    const searchParams = new URLSearchParams()
    if (params?.monitored !== undefined) searchParams.set("monitored", String(params.monitored))
    if (params?.status) searchParams.set("status", params.status)
    if (params?.search) searchParams.set("search", params.search)
    if (params?.sortBy) searchParams.set("sortBy", params.sortBy)
    if (params?.sortDir) searchParams.set("sortDir", params.sortDir)
    if (params?.limit) searchParams.set("limit", String(params.limit))
    if (params?.offset) searchParams.set("offset", String(params.offset))
    const query = searchParams.toString()
    return fetchAPI<PaginatedResponse<Author>>(`/author${query ? `?${query}` : ""}`)
  },

  get: (id: number) => fetchAPI<Author>(`/author/${id}`),

  getStats: (id: number) => fetchAPI<AuthorStats>(`/author/${id}/stats`),

  create: (data: CreateAuthorInput) =>
    fetchAPI<Author>("/author", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: UpdateAuthorInput) =>
    fetchAPI<Author>(`/author/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number, deleteFiles?: boolean) =>
    fetchAPI<void>(`/author/${id}${deleteFiles ? "?deleteFiles=true" : ""}`, {
      method: "DELETE",
    }),
}
