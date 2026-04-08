import { useState, useCallback, useEffect } from "react"
import { useMetadataSearchBooks, useDigitalLibrarySearch } from "../../hooks/useMetadata"
import { useSearch, type SearchParams } from "../../hooks/useIndexers"
import { Input, Button, Card, CardContent, Badge } from "../ui"
import type { MetadataBookResult, SearchResult, DigitalLibraryResult } from "../../lib/api"

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

interface LibraryResultProps {
  result: DigitalLibraryResult
}

function LibraryResult({ result }: LibraryResultProps) {
  const handleDownload = () => {
    window.open(result.downloadUrl || result.infoUrl, "_blank")
  }

  return (
    <Card>
      <CardContent className="py-3 px-4">
        <div className="flex justify-between items-center gap-4">
          <div className="flex-1 min-w-0">
            <h4 className="font-medium text-sm line-clamp-2">{result.title}</h4>
            {result.authors && result.authors.length > 0 && (
              <p className="text-xs text-muted-foreground">{result.authors.join(", ")}</p>
            )}
            <div className="flex flex-wrap gap-1 mt-1">
              <Badge variant="outline" className="text-xs">{result.provider}</Badge>
              <Badge variant="secondary" className="text-xs uppercase">{result.format}</Badge>
              {result.size > 0 && (
                <Badge variant="secondary" className="text-xs">{formatBytes(result.size)}</Badge>
              )}
              {result.year && (
                <Badge variant="secondary" className="text-xs">{result.year}</Badge>
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

interface SearchProgressProps {
  libraryLoading: boolean
  libraryError: Error | null
  libraryResults: number
  indexerLoading: boolean
  indexerError: Error | null
  indexerResults: number
  showCompact?: boolean
}

function SearchProgress({ 
  libraryLoading, 
  libraryError, 
  libraryResults,
  indexerLoading, 
  indexerError, 
  indexerResults,
  showCompact = false
}: SearchProgressProps) {
  const sources = [
    { 
      name: "Anna's Archive", 
      icon: "📚", 
      loading: libraryLoading, 
      error: libraryError,
      done: !libraryLoading 
    },
    { 
      name: "Library Genesis", 
      icon: "📖", 
      loading: libraryLoading, 
      error: libraryError,
      done: !libraryLoading 
    },
    { 
      name: "Internet Archive", 
      icon: "🏛️", 
      loading: libraryLoading, 
      error: libraryError,
      done: !libraryLoading 
    },
    { 
      name: "Torrent Indexers", 
      icon: "🔍", 
      loading: indexerLoading, 
      error: indexerError,
      done: !indexerLoading 
    },
  ]

  // If compact and no loading/errors, don't show
  if (showCompact && !libraryLoading && !indexerLoading && !libraryError && !indexerError) {
    return null
  }

  return (
    <div className="space-y-3 py-4">
      {sources.map((source, index) => (
        <div 
          key={source.name} 
          className="flex items-center gap-3 px-2"
          style={{ 
            animation: `fadeSlideIn 0.3s ease-out ${index * 0.1}s both` 
          }}
        >
          <span className="text-lg">{source.icon}</span>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">{source.name}</span>
              {source.loading && (
                <div className="h-3 w-3 border-2 border-primary border-t-transparent rounded-full animate-spin" />
              )}
              {source.done && !source.error && (
                <span className="text-green-500 text-xs">✓</span>
              )}
              {source.error && (
                <span className="text-destructive text-xs">✗</span>
              )}
            </div>
            <div className="h-1 bg-muted rounded-full overflow-hidden mt-1">
              <div 
                className={`h-full transition-all duration-500 ${
                  source.loading 
                    ? "bg-primary animate-pulse w-2/3" 
                    : source.error 
                      ? "bg-destructive w-full"
                      : "bg-green-500 w-full"
                }`}
              />
            </div>
          </div>
        </div>
      ))}
      
      <style>{`
        @keyframes fadeSlideIn {
          from {
            opacity: 0;
            transform: translateX(-10px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }
      `}</style>
      
      {(libraryResults > 0 || indexerResults > 0) && (libraryLoading || indexerLoading) && (
        <div className="text-center text-sm text-muted-foreground mt-4 pt-4 border-t">
          Found {libraryResults + indexerResults} results so far...
        </div>
      )}
    </div>
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
  
  // Search Torznab/Newznab indexers
  const indexerSearch = useSearch(searchParams)
  
  // Search digital libraries (Anna's Archive, LibGen, Internet Archive)
  const librarySearch = useDigitalLibrarySearch(searchQuery, true)

  const isLoading = indexerSearch.isLoading || librarySearch.isLoading
  const allSourcesFailed = indexerSearch.error && librarySearch.error
  const someSourcesFailed = (indexerSearch.error || librarySearch.error) && !allSourcesFailed

  // Filter indexer results that look like ebooks
  const indexerResults = (indexerSearch.data?.results ?? []).filter((result) => {
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

  // Digital library results (already filtered by format on backend)
  const libraryResults = librarySearch.data?.results ?? []

  const totalResults = indexerResults.length + libraryResults.length
  const hasResults = totalResults > 0

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div className="bg-background rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col" onClick={(e) => e.stopPropagation()}>
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
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {isLoading && (
            <SearchProgress
              libraryLoading={librarySearch.isLoading}
              libraryError={librarySearch.error}
              libraryResults={libraryResults.length}
              indexerLoading={indexerSearch.isLoading}
              indexerError={indexerSearch.error}
              indexerResults={indexerResults.length}
            />
          )}

          {/* Warning for partial failures - some sources failed but we have results */}
          {!isLoading && someSourcesFailed && hasResults && (
            <div className="bg-amber-500/10 border border-amber-500/30 rounded p-3 text-sm">
              <p className="text-amber-600 dark:text-amber-400 font-medium flex items-center gap-2">
                <span>⚠️</span> Some sources unavailable
              </p>
              <p className="text-amber-600/80 dark:text-amber-400/80 mt-1">
                {indexerSearch.error && "Torrent indexers (Prowlarr) not reachable. "}
                {librarySearch.error && "Digital libraries not responding. "}
                Showing results from available sources.
              </p>
            </div>
          )}

          {/* Error when ALL sources failed */}
          {!isLoading && allSourcesFailed && (
            <div className="bg-destructive/10 border border-destructive/30 rounded p-4">
              <p className="text-destructive font-medium">All sources failed</p>
              <p className="text-sm text-destructive/80 mt-1">
                Could not reach any download source. Check your network connection.
              </p>
            </div>
          )}

          {/* Digital Library Results - show even while indexer is loading */}
          {!librarySearch.isLoading && libraryResults.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold text-muted-foreground mb-2 flex items-center gap-2">
                <span>📚</span> Digital Libraries
                <Badge variant="secondary" className="text-xs">{libraryResults.length}</Badge>
              </h4>
              <div className="space-y-2">
                {libraryResults.map((result) => (
                  <LibraryResult key={`${result.provider}-${result.id}`} result={result} />
                ))}
              </div>
            </div>
          )}

          {/* Indexer Results - show even while library is loading */}
          {!indexerSearch.isLoading && indexerResults.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold text-muted-foreground mb-2 flex items-center gap-2">
                <span>🔍</span> Torrent/Usenet Indexers
                <Badge variant="secondary" className="text-xs">{indexerResults.length}</Badge>
              </h4>
              <div className="space-y-2">
                {indexerResults.map((result) => (
                  <DownloadResult key={result.guid} result={result} />
                ))}
              </div>
            </div>
          )}

          {/* No Results */}
          {!isLoading && !allSourcesFailed && totalResults === 0 && (
            <div className="text-center text-muted-foreground py-8">
              <p className="text-lg mb-2">No downloads found</p>
              <p className="text-sm mb-4">
                Could not find "{book.title}" in any library or indexer.
              </p>
              <div className="text-xs bg-muted/50 rounded p-3 text-left max-w-md mx-auto">
                <p className="font-medium mb-1">Sources checked:</p>
                <p>• Digital libraries (Anna's Archive, LibGen, Internet Archive)</p>
                <p>• Torrent/Usenet indexers (via Prowlarr)</p>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t text-sm text-muted-foreground">
          {totalResults > 0 && (
            <span>{totalResults} download {totalResults === 1 ? "option" : "options"} found</span>
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
