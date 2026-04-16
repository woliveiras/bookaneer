import { useNavigate } from "@tanstack/react-router"
import { useState } from "react"
import { BookCard } from "../../components/books/BookCard"
import { Badge, Button, Input } from "../../components/ui"
import { useBooks } from "../../hooks/useBooks"

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
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3 flex-1 min-w-0">
          <div className="w-full sm:w-1/2">
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
        </div>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {[...Array(8)].map((_, i) => (
            <div
              // biome-ignore lint/suspicious/noArrayIndexKey: Static skeleton placeholders have no unique data
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
          <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 list-none p-0 m-0">
            {data.records.map((book) => (
              <li key={book.id}>
                <BookCard
                  book={book}
                  onClick={() =>
                    navigate({ to: "/book/$bookId", params: { bookId: String(book.id) } })
                  }
                />
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  )
}
