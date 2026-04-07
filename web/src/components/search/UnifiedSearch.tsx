import { useState, useCallback, useEffect } from "react"
import { useMetadataSearchBooks } from "../../hooks/useMetadata"
import { useSearch, type SearchParams } from "../../hooks/useIndexers"
import { Input, Button, Card, CardContent, Badge } from "../ui"
import type { MetadataBookResult, SearchResult } from "../../lib/api"

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B"
  const k = 1024
  const sizes = ["B", "KB", "MB", "GB"]
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`
}

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
            <Badge variant="outline" className="text-xs">{book.provider}</Badge>
            {book.isbn13 && <Badge variant="secondary" className="text-xs">{book.isbn13}</Badge>}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

interface DownloadResultProps {
  result: SearchResult
}

function DownloadResult({ result }: DownloadResultProps) {
  const handleDownload = () => {
    window.open(result.downloadUrl, "_blank")
  }

  return (
    <Card>
      <CardContent className="py-3 px-4">
        <div className="flex justify-between items-center gap-4">
          <div className="flex-1 min-w-0">
            <h4 className="font-medium text-sm line-clamp-2">{result.title}</h4>
            <div className="flex flex-wrap gap-1 mt-1">
              <Badge variant="outline" className="text-xs">{result.indexerName}</Badge>
              <Badge variant="secondary" className="text-xs">{formatBytes(result.size)}</Badge>
              {result.seeders !== undefined && result.seeders > 0 && (
                <Badge variant="default" className="bg-green-600 text-xs">
                  {result.seeders} seeds
                </Badge>
              )}
            </div>
          </div>
          <Button size="sm" onClick={handleDownload}>
            Download
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

interface DownloadPanelProps {
  book: MetadataBookResult
  onClose: () => void
}

function DownloadPanel({ book, onClose }: DownloadPanelProps) {
  // Build search query from book metadata
  const searchQuery = [book.title, ...(book.authors || [])].join(" ")
  const searchParams: SearchParams = { q: searchQuery }
  
  const { data, isLoading, error } = useSearch(searchParams)

  // Filter results that look like ebooks (handle null/undefined results)
  const ebookResults = (data?.results ?? []).filter((result) => {
    const title = result.title.toLowerCase()
    const category = result.category?.toLowerCase() || ""
    return (
      title.includes("epub") ||
      title.includes("pdf") ||
      title.includes("mobi") ||
      title.includes("azw") ||
      title.includes("ebook") ||
      category.includes("ebook") ||
      category.includes("book") ||
      result.size < 500 * 1024 * 1024
    )
  })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-background rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="p-4 border-b flex items-start gap-4">
          {book.coverUrl && (
            <img
              src={book.coverUrl}
              alt={book.title}
              className="w-16 h-24 object-cover rounded shadow"
            />
          )}
          <div className="flex-1 min-w-0">
            <h3 className="font-bold text-lg line-clamp-2">{book.title}</h3>
            {book.authors && book.authors.length > 0 && (
              <p className="text-muted-foreground">{book.authors.join(", ")}</p>
            )}
            {book.publishedYear && (
              <p className="text-sm text-muted-foreground">{book.publishedYear}</p>
            )}
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            ✕
          </Button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {isLoading && (
            <div className="flex justify-center py-8">
              <div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
            </div>
          )}

          {error && (
            <div className="bg-destructive/10 border border-destructive/30 rounded p-4">
              <p className="text-destructive font-medium">Error searching indexers</p>
              <p className="text-sm text-destructive/80 mt-1">{error.message}</p>
            </div>
          )}

          {!isLoading && !error && ebookResults.length === 0 && (
            <div className="text-center text-muted-foreground py-8">
              <p className="text-lg mb-2">No ebook downloads found</p>
              <p className="text-sm mb-4">
                {data?.total === 0 
                  ? "The indexers returned no results for this search."
                  : `Found ${data?.total ?? 0} results, but none matched ebook filters (epub, pdf, mobi, etc.)`
                }
              </p>
              <div className="text-xs bg-muted/50 rounded p-3 text-left max-w-md mx-auto">
                <p className="font-medium mb-1">Debug info:</p>
                <p><span className="text-muted-foreground">Query sent:</span> {searchQuery}</p>
                <p><span className="text-muted-foreground">Total results:</span> {data?.total ?? 0}</p>
                <p><span className="text-muted-foreground">After ebook filter:</span> {ebookResults.length}</p>
              </div>
            </div>
          )}

          {ebookResults.map((result) => (
            <DownloadResult
              key={result.guid}
              result={result}
            />
          ))}
        </div>

        {/* Footer */}
        <div className="p-4 border-t text-sm text-muted-foreground">
          {ebookResults.length > 0 && (
            <span>{ebookResults.length} ebook {ebookResults.length === 1 ? "release" : "releases"} found</span>
          )}
        </div>
      </div>
    </div>
  )
}

export function UnifiedSearch() {
  // Read initial query from URL
  const [query, setQuery] = useState(() => {
    const params = new URLSearchParams(window.location.search)
    return params.get("q") || ""
  })
  const [submittedQuery, setSubmittedQuery] = useState(() => {
    const params = new URLSearchParams(window.location.search)
    return params.get("q") || ""
  })
  const [selectedBook, setSelectedBook] = useState<MetadataBookResult | null>(null)

  // Update URL when search is submitted
  const handleSearch = useCallback(() => {
    if (query.trim().length >= 2) {
      const trimmedQuery = query.trim()
      setSubmittedQuery(trimmedQuery)
      setSelectedBook(null)
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
            {bookSearch.data.results.length} books found — click one to find downloads
          </p>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {bookSearch.data.results.map((book) => (
              <BookCard
                key={`${book.provider}-${book.foreignId}`}
                book={book}
                onSelect={setSelectedBook}
                isSelected={selectedBook?.foreignId === book.foreignId}
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
            <p className="text-lg text-muted-foreground mb-2">
              Search for any book
            </p>
            <p className="text-sm text-muted-foreground">
              Enter a title, author name, or ISBN to find books and download them
            </p>
          </CardContent>
        </Card>
      )}

      {/* Download panel */}
      {selectedBook && (
        <DownloadPanel
          book={selectedBook}
          onClose={() => setSelectedBook(null)}
        />
      )}
    </div>
  )
}
