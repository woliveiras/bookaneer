import { useNavigate } from "@tanstack/react-router"
import { useCallback, useEffect, useState } from "react"
import { Badge, Button, Card, CardContent, Input } from "../../components/ui"
import { useMetadataSearchBooks } from "../../hooks/useMetadata"
import type { MetadataBookResult } from "../../lib/api"

interface BookCardProps {
  book: MetadataBookResult
  onSelect: (book: MetadataBookResult) => void
  isSelected: boolean
}

function BookCard({ book, onSelect, isSelected }: BookCardProps) {
  return (
    <Card
      className={`cursor-pointer transition-colors ${
        isSelected ? "border-primary ring-2 ring-primary" : "hover:border-primary"
      }`}
      onClick={() => onSelect(book)}
    >
      <CardContent className="p-4 flex gap-4">
        {book.coverUrl ? (
          <img
            src={book.coverUrl}
            alt={book.title}
            className="w-16 h-24 object-cover rounded shadow"
            loading="lazy"
          />
        ) : (
          <div className="w-16 h-24 bg-muted rounded flex items-center justify-center text-2xl">
            📚
          </div>
        )}
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold line-clamp-2">{book.title}</h3>
          {book.authors && book.authors.length > 0 && (
            <p className="text-sm text-muted-foreground">{book.authors.join(", ")}</p>
          )}
          {book.publishedYear && (
            <p className="text-sm text-muted-foreground">{book.publishedYear}</p>
          )}
          <div className="flex flex-wrap gap-1 mt-2">
            <Badge variant="outline" className="text-xs">
              {book.provider}
            </Badge>
            {book.isbn13 && (
              <Badge variant="secondary" className="text-xs">
                {book.isbn13}
              </Badge>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export function UnifiedSearch() {
  const navigate = useNavigate()

  // Read initial query from URL
  const [query, setQuery] = useState(() => {
    const params = new URLSearchParams(window.location.search)
    return params.get("q") || ""
  })
  const [submittedQuery, setSubmittedQuery] = useState(() => {
    const params = new URLSearchParams(window.location.search)
    return params.get("q") || ""
  })

  // Navigate to book details page
  const handleSelectBook = useCallback(
    (book: MetadataBookResult) => {
      navigate({
        to: "/search/book",
        search: {
          title: book.title,
          authors: book.authors?.join("|||"),
          provider: book.provider,
          foreignId: book.foreignId,
          publishedYear: book.publishedYear?.toString(),
          coverUrl: book.coverUrl,
          isbn13: book.isbn13,
        },
      })
    },
    [navigate],
  )

  // Update URL when search is submitted
  const handleSearch = useCallback(() => {
    if (query.trim().length >= 2) {
      const trimmedQuery = query.trim()
      setSubmittedQuery(trimmedQuery)
      // Update URL without navigation
      const url = new URL(window.location.href)
      url.searchParams.set("q", trimmedQuery)
      window.history.replaceState({}, "", url.toString())
    }
  }, [query])

  // Handle browser back/forward
  useEffect(() => {
    const handlePopState = () => {
      const params = new URLSearchParams(window.location.search)
      const q = params.get("q") || ""
      setQuery(q)
      setSubmittedQuery(q)
    }
    window.addEventListener("popstate", handlePopState)
    return () => window.removeEventListener("popstate", handlePopState)
  }, [])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        handleSearch()
      }
    },
    [handleSearch],
  )

  const bookSearch = useMetadataSearchBooks(submittedQuery, submittedQuery.length >= 2)

  return (
    <div className="space-y-6">
      {/* Search input */}
      <div className="flex gap-2">
        <div className="flex-1 relative">
          <Input
            type="search"
            placeholder="Search by book title, author, or ISBN..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            className="text-lg py-6"
            aria-label="Search for books"
          />
        </div>
        <Button
          onClick={handleSearch}
          disabled={query.trim().length < 2}
          size="lg"
          className="px-8"
        >
          Search
        </Button>
      </div>

      {/* Loading state */}
      {bookSearch.isLoading && (
        <div className="flex justify-center py-12">
          <div className="animate-spin h-10 w-10 border-4 border-primary border-t-transparent rounded-full" />
        </div>
      )}

      {/* No results */}
      {!bookSearch.isLoading && submittedQuery && bookSearch.data?.results.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-lg text-muted-foreground mb-2">
              No books found for "{submittedQuery}"
            </p>
            <p className="text-sm text-muted-foreground">
              Try a different spelling or search for the author's name
            </p>
          </CardContent>
        </Card>
      )}

      {/* Book results */}
      {bookSearch.data && bookSearch.data.results.length > 0 && (
        <div>
          <p className="text-sm text-muted-foreground mb-4">
            {bookSearch.data.results.length} books found — click one for manual search
          </p>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {bookSearch.data.results.map((book) => (
              <BookCard
                key={`${book.provider}-${book.foreignId}`}
                book={book}
                onSelect={handleSelectBook}
                isSelected={false}
              />
            ))}
          </div>
        </div>
      )}

      {/* Empty state */}
      {!submittedQuery && !bookSearch.isLoading && (
        <Card>
          <CardContent className="py-12 text-center">
            <div className="text-4xl mb-4">📚</div>
            <p className="text-lg text-muted-foreground mb-2">Search for any book</p>
            <p className="text-sm text-muted-foreground">
              Enter a title, author name, or ISBN to find books and download them
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
