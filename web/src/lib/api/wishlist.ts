import type { Book, PaginatedResponse } from "../schemas"
import { fetchAPI } from "./client"

export const wishlistApi = {
  list: () => fetchAPI<PaginatedResponse<Book>>("/wishlist"),

  add: (data: {
    title: string
    authors: string[]
    foreignId?: string
    isbn13?: string
    imageUrl?: string
  }) =>
    fetchAPI<Book>("/wishlist", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  remove: (id: number) =>
    fetchAPI<void>(`/wishlist/${id}`, {
      method: "DELETE",
    }),
}
