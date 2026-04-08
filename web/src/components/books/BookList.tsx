import { useState } from "react"
import { useNavigate } from "@tanstack/react-router"
import { useBooks } from "../../hooks/useBooks"
import { BookCard } from "./BookCard"
import { Input, Button, Badge } from "../ui"

interface BookListProps {
  authorId?: number
}

export function BookList({ authorId }: BookListProps) {
  const navigate = useNavigate()
  const [search, setSearch] = useState("")
  const [debouncedSearch, setDebouncedSearch] = useState("")
  const [showMissing, setShowMissing] = useState(false)

  const { data, isLoading, error } = useBooks({
    search: debouncedSearch || undefined,
    authorId,
    missing: showMissing || undefined,
    limit: 50,
  })

  // Simple debounce
  const handleSearch = (value: string) => {
    setSearch(value)
    const timeoutId = setTimeout(() => {
      setDebouncedSearch(value)
    }, 300)
    return () => clearTimeout(timeoutId)
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive bg-destructive/10 p-4" role="alert">
        <p className="text-destructive">Failed to load books: {error.message}</p>
        <Button variant="outline" className="mt-2" onClick={() => window.location.reload()}>
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="w-64">
            <label htmlFor="book-search" className="sr-only">
              Search books
            </label>
            <Input
              id="book-search"
              type="search"
              placeholder="Search by title or ISBN..."
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
            />
          </div>
          <Button
            variant={showMissing ? "default" : "outline"}
            size="sm"
            onClick={() => setShowMissing(!showMissing)}
            aria-pressed={showMissing}
          >
            Missing Only
          </Button>
        </div>
        <Button>Add Book</Button>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {[...Array(8)].map((_, i) => (
            <div
              key={i}
              className="h-36 animate-pulse rounded-lg border bg-muted"
              aria-hidden="true"
            />
          ))}
        </div>
      ) : !data?.records?.length ? (
        <div className="rounded-lg border border-dashed p-8 text-center">
          <p className="text-muted-foreground">
            {debouncedSearch
              ? `No books found for "${debouncedSearch}"`
              : showMissing
                ? "No missing books"
                : "No books yet"}
          </p>
          {!debouncedSearch && !showMissing && (
            <Button variant="link" className="mt-2">
              Add your first book
            </Button>
          )}
        </div>
      ) : (
        <>
          <div className="flex items-center gap-2">
            <p className="text-sm text-muted-foreground">
              {data.totalRecords} book{data.totalRecords !== 1 ? "s" : ""}
            </p>
            {showMissing && <Badge variant="outline">Missing only</Badge>}
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" role="list">
            {data.records.map((book) => (
              <div key={book.id} role="listitem">
                <BookCard
                  book={book}
                  onClick={() => navigate({ to: "/book/$bookId", params: { bookId: String(book.id) } })}
                />
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
