import { useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { Button, Card, CardContent } from "../../components/ui"
import { useUpdateBook } from "../../hooks/useBooks"
import { useManualGrab, useSearchBook, useWantedMissing } from "../../hooks/useWanted"
import { BookOpen, PartyPopper } from "lucide-react"
import type { Book } from "../../lib/api"
import type { BookSearchResult } from "../../lib/api"

export function WantedList() {
  const queryClient = useQueryClient()
  const { data, isLoading, error, refetch } = useWantedMissing()
  const searchBookMutation = useSearchBook()
  const manualGrabMutation = useManualGrab()
  const updateBookMutation = useUpdateBook()
  const [searchingBooks, setSearchingBooks] = useState<Set<number>>(new Set())
  const [removingBooks, setRemovingBooks] = useState<Set<number>>(new Set())
  const [bookToRemove, setBookToRemove] = useState<Book | null>(null)
  const [searchState, setSearchState] = useState<{
    bookId: number
    bookTitle: string
    results: BookSearchResult[]
    noResults: boolean
  } | null>(null)

  const handleSearchBook = async (bookId: number, bookTitle: string) => {
    setSearchingBooks((prev) => new Set(prev).add(bookId))
    try {
      const response = await searchBookMutation.mutateAsync(bookId)
      setSearchState({ bookId, bookTitle, results: response.results, noResults: response.noResults })
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

  const handleGrab = async (result: BookSearchResult) => {
    if (!searchState) return
    try {
      await manualGrabMutation.mutateAsync({
        bookId: searchState.bookId,
        downloadUrl: result.downloadUrl,
        releaseTitle: result.title,
        size: result.size,
      })
      queryClient.invalidateQueries({ queryKey: ["queue"] })
      queryClient.invalidateQueries({ queryKey: ["wanted"] })
      setSearchState(null)
    } catch (err) {
      console.error("Failed to grab:", err)
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
        </div>
      </div>

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
              onSearch={() => handleSearchBook(book.id, book.title)}
              isSearching={searchingBooks.has(book.id)}
              onRemove={() => handleRemoveFromWanted(book)}
              isRemoving={removingBooks.has(book.id)}
            />
          ))}
        </div>
      )}

      {/* Search results modal */}
      {searchState && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background p-6 rounded-lg shadow-lg max-w-2xl w-full mx-4 border max-h-[80vh] flex flex-col">
            <h3 className="text-lg font-semibold mb-1">Search Results</h3>
            <p className="text-sm text-muted-foreground mb-4">"{searchState.bookTitle}"</p>
            {searchState.noResults ? (
              <div className="py-6 text-center">
                <p className="text-muted-foreground mb-4">No results found. This book remains in your Wanted list.</p>
                <Button variant="outline" onClick={() => setSearchState(null)}>Close</Button>
              </div>
            ) : (
              <>
                <div className="overflow-y-auto flex-1 space-y-2 mb-4">
                  {searchState.results.map((result, i) => (
                    <div key={i} className="border rounded-lg p-3 flex items-center justify-between gap-3">
                      <div className="min-w-0 flex-1">
                        <p className="font-medium text-sm truncate">{result.title}</p>
                        <div className="flex flex-wrap gap-2 mt-1">
                          <span className="text-xs bg-muted px-2 py-0.5 rounded">{result.format.toUpperCase()}</span>
                          <span className="text-xs text-muted-foreground">{(result.size / 1024 / 1024).toFixed(1)} MB</span>
                          <span className="text-xs text-muted-foreground">{result.provider}</span>
                          {result.seeders !== undefined && result.seeders > 0 && (
                            <span className="text-xs text-muted-foreground">{result.seeders} seeders</span>
                          )}
                        </div>
                      </div>
                      <Button
                        size="sm"
                        onClick={() => handleGrab(result)}
                        disabled={manualGrabMutation.isPending}
                      >
                        Grab
                      </Button>
                    </div>
                  ))}
                </div>
                <div className="flex justify-end">
                  <Button variant="outline" onClick={() => setSearchState(null)}>Cancel</Button>
                </div>
              </>
            )}
          </div>
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
