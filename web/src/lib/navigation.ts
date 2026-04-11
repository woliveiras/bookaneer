import type { NavigateFn } from "@tanstack/react-router"
import type { Book } from "./types/book"

/**
 * Navigates to the book search page pre-filled with the given book's metadata.
 * Used when the user wants to manually search for a download of a specific book.
 */
export function navigateToBookSearch(navigate: NavigateFn, book: Book): void {
  navigate({
    to: "/search/book",
    search: {
      title: book.title,
      authors: book.authorName,
      foreignId: book.foreignId || undefined,
      isbn13: book.isbn13 || undefined,
      coverUrl: book.imageUrl || undefined,
      publishedYear: book.releaseDate
        ? String(new Date(book.releaseDate).getFullYear())
        : undefined,
      autoSearch: "true",
      bookId: String(book.id),
    },
  })
}
