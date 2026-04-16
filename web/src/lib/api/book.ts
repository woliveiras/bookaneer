import type {
  Book,
  BookWithEditions,
  CreateBookInput,
  ListBooksParams,
  PaginatedResponse,
} from "../types"
import { fetchAPI } from "./client"

export const bookApi = {
  list: (params?: ListBooksParams) => {
    const searchParams = new URLSearchParams()
    if (params?.authorId) searchParams.set("authorId", String(params.authorId))
    if (params?.monitored !== undefined) searchParams.set("monitored", String(params.monitored))
    if (params?.missing) searchParams.set("missing", "true")
    if (params?.inWishlist) searchParams.set("in_wishlist", "true")
    if (params?.search) searchParams.set("search", params.search)
    if (params?.sortBy) searchParams.set("sortBy", params.sortBy)
    if (params?.sortDir) searchParams.set("sortDir", params.sortDir)
    if (params?.limit) searchParams.set("limit", String(params.limit))
    if (params?.offset) searchParams.set("offset", String(params.offset))
    const query = searchParams.toString()
    return fetchAPI<PaginatedResponse<Book>>(`/book${query ? `?${query}` : ""}`)
  },

  get: (id: number) => fetchAPI<BookWithEditions>(`/book/${id}`),

  create: (data: CreateBookInput) =>
    fetchAPI<Book>("/book", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  update: (id: number, data: Partial<CreateBookInput>) =>
    fetchAPI<Book>(`/book/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  delete: (id: number) =>
    fetchAPI<void>(`/book/${id}`, {
      method: "DELETE",
    }),

  addToWishlist: (data: { title: string; authors: string[]; foreignId?: string; isbn13?: string; imageUrl?: string }) =>
    fetchAPI<Book>("/wishlist", {
      method: "POST",
      body: JSON.stringify(data),
    }),
}
