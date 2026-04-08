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

// Pirate ship SVG component
function PirateShip({ className }: { className?: string }) {
  return (
    <svg 
      viewBox="0 0 64 48" 
      className={className}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* Hull */}
      <path 
        d="M8 32 L12 40 L52 40 L56 32 L48 32 L48 28 L16 28 L16 32 Z" 
        fill="currentColor" 
        className="text-amber-800"
      />
      {/* Deck */}
      <rect x="16" y="24" width="32" height="4" fill="currentColor" className="text-amber-700" />
      {/* Mast */}
      <rect x="30" y="4" width="4" height="24" fill="currentColor" className="text-amber-900" />
      {/* Sail */}
      <path 
        d="M34 6 L34 22 L50 22 Q42 14 34 6 Z" 
        fill="currentColor" 
        className="text-slate-100"
      />
      {/* Flag - Black */}
      <rect x="31" y="0" width="12" height="8" fill="currentColor" className="text-slate-800" />
    </svg>
  )
}

// Animated waves SVG
function WavesSVG() {
  return (
    <svg 
      viewBox="0 0 400 20" 
      className="absolute bottom-0 left-0 w-[200%] h-5"
      preserveAspectRatio="none"
    >
      <path 
        d="M0 10 Q25 0 50 10 T100 10 T150 10 T200 10 T250 10 T300 10 T350 10 T400 10 V20 H0 Z" 
        fill="currentColor" 
        className="text-blue-400/40"
      />
      <path 
        d="M0 14 Q25 8 50 14 T100 14 T150 14 T200 14 T250 14 T300 14 T350 14 T400 14 V20 H0 Z" 
        fill="currentColor" 
        className="text-blue-500/50"
      />
    </svg>
  )
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
  const isLoading = libraryLoading || indexerLoading
  const totalFound = libraryResults + indexerResults
  
  // If compact and no loading, don't show
  if (showCompact && !isLoading) {
    return null
  }

  // Source status for the legend
  const sources = [
    { name: "Anna's Archive", done: !libraryLoading, error: libraryError },
    { name: "Library Genesis", done: !libraryLoading, error: libraryError },
    { name: "Internet Archive", done: !libraryLoading, error: libraryError },
    { name: "Torrent Indexers", done: !indexerLoading, error: indexerError },
  ]

  return (
    <div className="py-6">
      {/* Pirate Ship Animation */}
      <div className="relative h-24 mx-auto max-w-sm overflow-hidden rounded-lg">
        {/* Islands/Ports */}
        <div className="absolute bottom-4 left-2 text-2xl" title="Library">
          📚
        </div>
        <div
          className="absolute bottom-4 right-2 text-2xl"
          title="Your Collection"
        >
          🌍
        </div>

        {/* Ship */}
        <div
          className={`absolute bottom-3 w-16 h-12 ${isLoading ? "animate-sail" : "left-1/2 -translate-x-1/2"}`}
          style={{
            transition: isLoading ? undefined : "all 0.5s ease-out",
          }}
        >
          <div className="animate-bob">
            <PirateShip className="w-full h-full drop-shadow-md" />
          </div>
        </div>

        {/* Waves */}
        <div
          className={`absolute bottom-0 left-0 w-full overflow-hidden ${isLoading ? "animate-waves" : ""}`}
        >
          <WavesSVG />
        </div>
      </div>

      {/* Status text */}
      <div className="text-center mt-4">
        {isLoading ? (
          <p className="text-sm text-muted-foreground animate-pulse">
            Sailing the seven seas for books...
          </p>
        ) : totalFound > 0 ? (
          <p className="text-sm text-green-600 dark:text-green-400">
            Found {totalFound} treasure{totalFound !== 1 ? "s" : ""}!
          </p>
        ) : (
          <p className="text-sm text-muted-foreground">
            No treasures found in these waters
          </p>
        )}
      </div>

      {/* Source indicators - compact */}
      <div className="flex justify-center gap-3 mt-3 flex-wrap">
        {sources.map((source) => (
          <div
            key={source.name}
            className="flex items-center gap-1.5 text-xs text-muted-foreground"
          >
            {source.done ? (
              source.error ? (
                <span className="text-destructive">✗</span>
              ) : (
                <span className="text-green-500">✓</span>
              )
            ) : (
              <div className="relative h-3.5 w-3.5">
                <div
                  className="absolute inset-0 rounded-full animate-gradient-spin"
                  style={{
                    background:
                      "conic-gradient(from 0deg, transparent, #60a5fa, #3b82f6, transparent)",
                  }}
                />
                <div className="absolute inset-0.5 rounded-full bg-background" />
              </div>
            )}
            <span className="hidden sm:inline">{source.name}</span>
          </div>
        ))}
      </div>

      <style>{`
        @keyframes sail {
          0%, 100% { left: 10%; }
          50% { left: calc(90% - 4rem); }
        }
        @keyframes bob {
          0%, 100% { transform: translateY(0) rotate(-2deg); }
          25% { transform: translateY(-3px) rotate(0deg); }
          50% { transform: translateY(0) rotate(2deg); }
          75% { transform: translateY(2px) rotate(0deg); }
        }
        @keyframes waves {
          0% { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
        @keyframes gradient-spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
        .animate-sail {
          animation: sail 4s ease-in-out infinite;
        }
        .animate-bob {
          animation: bob 2s ease-in-out infinite;
        }
        .animate-gradient-spin {
          animation: gradient-spin 1s linear infinite;
        }
        .animate-waves {
          animation: waves 3s linear infinite;
        }
      `}</style>
    </div>
  );
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
