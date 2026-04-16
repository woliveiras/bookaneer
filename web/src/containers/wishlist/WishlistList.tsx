import { useNavigate } from "@tanstack/react-router"
import { PartyPopper } from "lucide-react"
import { useState } from "react"
import { BookCard } from "../../components/books/BookCard"
import { Button, Card, CardContent } from "../../components/ui"
import { useRemoveFromWishlist, useWishlist } from "../../hooks/useWishlist"
import type { Book } from "../../lib/api"

export function WishlistList() {
  const navigate = useNavigate()
  const { data, isLoading, error, refetch } = useWishlist()
  const removeFromWishlist = useRemoveFromWishlist()
  const [removingBooks, setRemovingBooks] = useState<Set<number>>(new Set())
  const [bookToRemove, setBookToRemove] = useState<Book | null>(null)

  const handleSearchBook = (book: Book) => {
    navigate({
      to: "/search/releases",
      search: {
        title: book.title,
        authors: book.authorName ?? "",
        publishedYear: book.releaseDate
          ? String(new Date(book.releaseDate).getFullYear())
          : undefined,
        coverUrl: book.imageUrl ?? "",
        foreignId: "",
        isbn13: book.isbn13 ?? "",
        bookId: book.id ? String(book.id) : undefined,
      },
    })
  }

  const confirmRemove = async () => {
    if (!bookToRemove) return
    const bookId = bookToRemove.id
    setRemovingBooks((prev) => new Set(prev).add(bookId))
    try {
      await removeFromWishlist.mutateAsync(bookId)
      setBookToRemove(null)
    } catch (err) {
      console.error("Failed to remove from wishlist:", err)
    } finally {
      setRemovingBooks((prev) => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    )
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <p className="text-destructive">Failed to load wishlist</p>
          <Button variant="outline" onClick={() => refetch()} className="mt-4">
            Try Again
          </Button>
        </CardContent>
      </Card>
    )
  }

  const books = data?.records ?? []

  return (
    <div className="space-y-6">
      {/* Header with actions */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-muted-foreground">
            {books.length} {books.length === 1 ? "book" : "books"} in your wishlist
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => refetch()} disabled={isLoading}>
            Refresh
          </Button>
        </div>
      </div>

      {/* Empty state */}
      {books.length === 0 && (
        <Card>
          <CardContent className="p-12 text-center">
            <div className="flex justify-center mb-4">
              <PartyPopper className="w-8 h-8 text-muted-foreground" />
            </div>
            <h3 className="text-lg font-semibold mb-2">All caught up!</h3>
            <p className="text-muted-foreground">
              Your wishlist is empty. Search for books and click the bookmark icon to add them.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Books list */}
      {books.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {books.map((book) => (
            <BookCard
              key={book.id}
              book={book}
              actions={
                <>
                  <Button
                    size="sm"
                    onClick={() => handleSearchBook(book)}
                    disabled={removingBooks.has(book.id)}
                  >
                    Search
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setBookToRemove(book)}
                    disabled={removingBooks.has(book.id)}
                  >
                    {removingBooks.has(book.id) ? "Removing..." : "Remove"}
                  </Button>
                </>
              }
            />
          ))}
        </div>
      )}

      {/* Remove confirmation modal */}
      {bookToRemove && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4 border">
            <h3 className="text-lg font-semibold mb-2">Remove from Wishlist?</h3>
            <p className="text-muted-foreground mb-4">
              Are you sure you want to remove "{bookToRemove.title}" from your wishlist?
            </p>
            <div className="flex gap-2 justify-end">
              <Button variant="outline" onClick={() => setBookToRemove(null)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={confirmRemove}
                disabled={removingBooks.has(bookToRemove.id)}
              >
                {removingBooks.has(bookToRemove.id) ? "Removing..." : "Remove"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
