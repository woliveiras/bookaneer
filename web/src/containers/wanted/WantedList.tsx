import { useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { Button, Card, CardContent } from "../../components/ui"
import { useUpdateBook } from "../../hooks/useBooks"
import { useSearchAllMissing, useSearchBook, useWantedMissing } from "../../hooks/useWanted"
import { BookOpen, PartyPopper } from "lucide-react"
import type { Book } from "../../lib/api"

export function WantedList() {
  const queryClient = useQueryClient()
  const { data, isLoading, error, refetch } = useWantedMissing()
  const searchAllMutation = useSearchAllMissing()
  const searchBookMutation = useSearchBook()
  const updateBookMutation = useUpdateBook()
  const [searchingBooks, setSearchingBooks] = useState<Set<number>>(new Set())
  const [removingBooks, setRemovingBooks] = useState<Set<number>>(new Set())
  const [bookToRemove, setBookToRemove] = useState<Book | null>(null)

  const handleSearchAll = async () => {
    try {
      await searchAllMutation.mutateAsync()
    } catch (err) {
      console.error("Failed to start search:", err)
    }
  }

  const handleSearchBook = async (bookId: number) => {
    setSearchingBooks((prev) => new Set(prev).add(bookId))
    try {
      await searchBookMutation.mutateAsync(bookId)
    } catch (err) {
      console.error("Failed to search book:", err)
    } finally {
      setSearchingBooks((prev) => {
        const next = new Set(prev)
        next.delete(bookId)
        return next
      })
    }
  }

  const handleRemoveFromWanted = (book: Book) => {
    setBookToRemove(book)
  }

  const confirmRemove = async () => {
    if (!bookToRemove) return
    const bookId = bookToRemove.id
    setRemovingBooks((prev) => new Set(prev).add(bookId))
    try {
      await updateBookMutation.mutateAsync({ id: bookId, data: { monitored: false } })
      queryClient.invalidateQueries({ queryKey: ["wanted"] })
      setBookToRemove(null)
    } catch (err) {
      console.error("Failed to remove from wanted:", err)
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
          <p className="text-destructive">Failed to load wanted books</p>
          <Button variant="outline" onClick={() => refetch()} className="mt-4">
            Try Again
          </Button>
        </CardContent>
      </Card>
    )
  }

  const books = data?.records || []

  return (
    <div className="space-y-6">
      {/* Header with actions */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-muted-foreground">
            {books.length} monitored {books.length === 1 ? "book" : "books"} missing from library
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => refetch()} disabled={isLoading}>
            Refresh
          </Button>
          <Button
            onClick={handleSearchAll}
            disabled={searchAllMutation.isPending || books.length === 0}
          >
            {searchAllMutation.isPending ? "Searching..." : "Search All"}
          </Button>
        </div>
      </div>

      {/* Success message */}
      {searchAllMutation.isSuccess && (
        <Card className="border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950">
          <CardContent className="p-4">
            <p className="text-green-700 dark:text-green-300">
              Search started for all missing books. Check the Activity tab.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Empty state */}
      {books.length === 0 && (
        <Card>
          <CardContent className="p-12 text-center">
            <div className="flex justify-center mb-4"><PartyPopper className="w-8 h-8 text-muted-foreground" /></div>
            <h3 className="text-lg font-semibold mb-2">All caught up!</h3>
            <p className="text-muted-foreground">
              No monitored books are missing from your library.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Books list */}
      {books.length > 0 && (
        <div className="grid gap-4">
          {books.map((book) => (
            <WantedBookCard
              key={book.id}
              book={book}
              onSearch={() => handleSearchBook(book.id)}
              isSearching={searchingBooks.has(book.id)}
              onRemove={() => handleRemoveFromWanted(book)}
              isRemoving={removingBooks.has(book.id)}
            />
          ))}
        </div>
      )}

      {/* Remove confirmation modal */}
      {bookToRemove && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4 border">
            <h3 className="text-lg font-semibold mb-2">Remove from Wanted?</h3>
            <p className="text-muted-foreground mb-4">
              Are you sure you want to stop monitoring "{bookToRemove.title}"? This will remove it
              from the Wanted list.
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

interface WantedBookCardProps {
  book: Book
  onSearch: () => void
  isSearching: boolean
  onRemove: () => void
  isRemoving: boolean
}

function WantedBookCard({
  book,
  onSearch,
  isSearching,
  onRemove,
  isRemoving,
}: WantedBookCardProps) {
  return (
    <Card>
      <div className="flex items-start gap-4 p-4">
        {/* Cover image */}
        <div className="shrink-0 w-16 h-24 bg-muted rounded overflow-hidden">
          {book.imageUrl ? (
            <img src={book.imageUrl} alt={book.title} className="w-full h-full object-cover" />
          ) : (
            <div className="w-full h-full flex items-center justify-center"><BookOpen className="w-6 h-6 text-muted-foreground" /></div>
          )}
        </div>

        {/* Book info + actions */}
        <div className="flex-1 min-w-0 space-y-3">
          <div>
            <h3 className="font-semibold truncate">{book.title}</h3>
            {book.authorName && <p className="text-sm text-muted-foreground">{book.authorName}</p>}
            <div className="flex flex-wrap gap-2 mt-2">
              {book.releaseDate && (
                <span className="text-xs bg-muted px-2 py-1 rounded">
                  {new Date(book.releaseDate).getFullYear()}
                </span>
              )}
              {book.isbn13 && (
                <span className="text-xs bg-muted px-2 py-1 rounded">ISBN: {book.isbn13}</span>
              )}
              <span className="text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 px-2 py-1 rounded">
                Missing
              </span>
            </div>
          </div>

          {/* Actions — sit below info, always fully visible */}
          <div className="flex gap-2">
            <Button size="sm" onClick={onSearch} disabled={isSearching || isRemoving}>
              {isSearching ? "Searching..." : "Search"}
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={onRemove}
              disabled={isRemoving || isSearching}
            >
              {isRemoving ? "Removing..." : "Remove"}
            </Button>
          </div>
        </div>
      </div>
    </Card>
  )
}
