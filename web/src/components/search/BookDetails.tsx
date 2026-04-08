import { useState, useMemo } from "react"
import { useNavigate } from "@tanstack/react-router"
import { useDigitalLibrarySearch } from "../../hooks/useMetadata"
import { useSearch, type SearchParams } from "../../hooks/useIndexers"
import { useCreateBook } from "../../hooks/useBooks"
import { useCreateAuthor, useAuthors } from "../../hooks/useAuthors"
import { Button, Card, CardContent, Badge, Input } from "../ui"
import type { MetadataBookResult, SearchResult, DigitalLibraryResult } from "../../lib/api"

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B"
  const k = 1024
  const sizes = ["B", "KB", "MB", "GB"]
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`
}

// Format filter options
const FORMAT_OPTIONS = [
  { value: "all", label: "All Formats" },
  { value: "epub", label: "EPUB" },
  { value: "pdf", label: "PDF" },
  { value: "mobi", label: "MOBI" },
] as const

// Provider filter options
const PROVIDER_OPTIONS = [
  { value: "all", label: "All Sources" },
  { value: "internet-archive", label: "Internet Archive" },
  { value: "libgen", label: "LibGen" },
  { value: "annas-archive", label: "Anna's Archive" },
  { value: "torrent", label: "Torrent Indexers" },
] as const

// Sort options
const SORT_OPTIONS = [
  { value: "score", label: "Best Match" },
  { value: "year-desc", label: "Newest First" },
  { value: "year-asc", label: "Oldest First" },
  { value: "format", label: "By Format" },
] as const

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
              {result.score && (
                <Badge variant="default" className="text-xs bg-primary">
                  Score: {result.score}
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

// Pirate ship SVG component
function PirateShip({ className }: { className?: string }) {
  return (
    <svg 
      viewBox="0 0 64 48" 
      className={className}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path 
        d="M8 32 L12 40 L52 40 L56 32 L48 32 L48 28 L16 28 L16 32 Z" 
        fill="currentColor" 
        className="text-amber-800"
      />
      <rect x="16" y="24" width="32" height="4" fill="currentColor" className="text-amber-700" />
      <rect x="30" y="4" width="4" height="24" fill="currentColor" className="text-amber-900" />
      <path 
        d="M34 6 L34 22 L50 22 Q42 14 34 6 Z" 
        fill="currentColor" 
        className="text-slate-100"
      />
      <rect x="31" y="0" width="12" height="8" fill="currentColor" className="text-slate-800" />
    </svg>
  )
}

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

interface BookDetailsProps {
  book: MetadataBookResult
}

export function BookDetails({ book }: BookDetailsProps) {
  const navigate = useNavigate()
  
  // Search state - only start when user clicks
  const [searchStarted, setSearchStarted] = useState(false)
  
  // Add to library state
  const [addedToLibrary, setAddedToLibrary] = useState(false)
  const [addingToLibrary, setAddingToLibrary] = useState(false)
  const [addError, setAddError] = useState<string | null>(null)
  
  const createBook = useCreateBook()
  const createAuthor = useCreateAuthor()
  const authorName = book.authors?.[0] || "Unknown Author"
  const { data: existingAuthors } = useAuthors({ search: authorName, limit: 1 })
  
  // Filters state
  const [formatFilter, setFormatFilter] = useState<string>("all")
  const [providerFilter, setProviderFilter] = useState<string>("all")
  const [sortBy, setSortBy] = useState<string>("score")
  const [searchInResults, setSearchInResults] = useState("")

  // Handle add to library
  const handleAddToLibrary = async () => {
    setAddingToLibrary(true)
    setAddError(null)
    
    try {
      let authorId: number
      
      // Check if author already exists
      if (existingAuthors?.records?.length && existingAuthors.records[0].name.toLowerCase() === authorName.toLowerCase()) {
        authorId = existingAuthors.records[0].id
      } else {
        // Create author
        const author = await createAuthor.mutateAsync({
          name: authorName,
          monitored: true,
        })
        authorId = author.id
      }
      
      // Create book
      await createBook.mutateAsync({
        authorId,
        title: book.title,
        foreignId: book.foreignId || "",
        isbn13: book.isbn13 || "",
        releaseDate: book.publishedYear ? `${book.publishedYear}-01-01` : "",
        imageUrl: book.coverUrl || "",
        monitored: true,
      })
      
      setAddedToLibrary(true)
    } catch (err) {
      setAddError(err instanceof Error ? err.message : "Failed to add to library")
    } finally {
      setAddingToLibrary(false)
    }
  }

  // Build search queries
  const librarySearchQuery = book.title
  const indexerSearchQuery = [book.title, ...(book.authors || [])].join(" ")
  const searchParams: SearchParams = { q: indexerSearchQuery }

  // Search queries - only enabled when searchStarted is true
  const indexerSearch = useSearch(searchStarted ? searchParams : { q: "" }, searchStarted)
  const librarySearch = useDigitalLibrarySearch(librarySearchQuery, searchStarted)

  // Loading and error states
  const libraryDone = !librarySearch.isLoading
  const indexerDone = !indexerSearch.isLoading
  const isLoading = !libraryDone || !indexerDone

  const libraryFailed = libraryDone && !!librarySearch.error && !librarySearch.data?.results?.length
  const indexerFailed = indexerDone && !!indexerSearch.error && !indexerSearch.data?.results?.length
  const someSourcesFailed = (libraryFailed || indexerFailed) && !(libraryFailed && indexerFailed)

  // Filter indexer results for ebooks
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

  const libraryResults = librarySearch.data?.results ?? []

  // Apply filters and sorting
  const filteredLibraryResults = useMemo(() => {
    let results = [...libraryResults]

    // Format filter
    if (formatFilter !== "all") {
      results = results.filter(r => r.format.toLowerCase() === formatFilter)
    }

    // Provider filter
    if (providerFilter !== "all" && providerFilter !== "torrent") {
      results = results.filter(r => r.provider === providerFilter)
    }

    // Text search in results
    if (searchInResults.trim()) {
      const search = searchInResults.toLowerCase()
      results = results.filter(r => 
        r.title.toLowerCase().includes(search) ||
        r.authors?.some(a => a.toLowerCase().includes(search))
      )
    }

    // Sorting
    switch (sortBy) {
      case "year-desc":
        results.sort((a, b) => (b.year || 0) - (a.year || 0))
        break
      case "year-asc":
        results.sort((a, b) => (a.year || 0) - (b.year || 0))
        break
      case "format":
        const formatOrder = { epub: 1, pdf: 2, mobi: 3 }
        results.sort((a, b) => {
          const aOrder = formatOrder[a.format.toLowerCase() as keyof typeof formatOrder] || 99
          const bOrder = formatOrder[b.format.toLowerCase() as keyof typeof formatOrder] || 99
          return aOrder - bOrder
        })
        break
      case "score":
      default:
        results.sort((a, b) => (b.score || 0) - (a.score || 0))
    }

    return results
  }, [libraryResults, formatFilter, providerFilter, sortBy, searchInResults])

  const filteredIndexerResults = useMemo(() => {
    if (providerFilter !== "all" && providerFilter !== "torrent") {
      return []
    }

    let results = [...indexerResults]

    // Text search
    if (searchInResults.trim()) {
      const search = searchInResults.toLowerCase()
      results = results.filter(r => r.title.toLowerCase().includes(search))
    }

    return results
  }, [indexerResults, providerFilter, searchInResults])

  const totalResults = filteredLibraryResults.length + filteredIndexerResults.length

  // Source status for loading indicator
  const sources = [
    { name: "Anna's Archive", done: libraryDone, error: librarySearch.error, retrying: (librarySearch.failureCount || 0) > 0 && librarySearch.isLoading },
    { name: "Library Genesis", done: libraryDone, error: librarySearch.error, retrying: (librarySearch.failureCount || 0) > 0 && librarySearch.isLoading },
    { name: "Internet Archive", done: libraryDone, error: librarySearch.error, retrying: (librarySearch.failureCount || 0) > 0 && librarySearch.isLoading },
    { name: "Torrent Indexers", done: indexerDone, error: indexerSearch.error, retrying: (indexerSearch.failureCount || 0) > 0 && indexerSearch.isLoading },
  ]

  return (
    <div className="space-y-6">
      {/* Book Header */}
      <div className="flex gap-6 p-6 rounded-lg border bg-card">
        {book.coverUrl ? (
          <img
            src={book.coverUrl}
            alt={book.title}
            className="w-32 h-48 object-cover rounded shadow-lg"
          />
        ) : (
          <div className="w-32 h-48 bg-muted rounded flex items-center justify-center text-4xl">
            📚
          </div>
        )}
        <div className="flex-1">
          <h1 className="text-2xl font-bold">{book.title}</h1>
          {book.authors && book.authors.length > 0 && (
            <p className="text-lg text-muted-foreground mt-1">
              {book.authors.join(", ")}
            </p>
          )}
          {book.publishedYear && (
            <p className="text-muted-foreground">{book.publishedYear}</p>
          )}
          <div className="flex flex-wrap gap-2 mt-4">
            <Badge variant="outline">{book.provider}</Badge>
            {book.isbn13 && <Badge variant="secondary">{book.isbn13}</Badge>}
          </div>
          <div className="flex flex-wrap gap-2 mt-4">
            <Button 
              variant="outline" 
              size="sm"
              onClick={() => navigate({ to: "/search" })}
            >
              ← Back to Search
            </Button>
            {!addedToLibrary ? (
              <Button 
                size="sm"
                variant="default"
                onClick={handleAddToLibrary}
                disabled={addingToLibrary}
              >
                {addingToLibrary ? "Adding..." : "➕ Add to Library"}
              </Button>
            ) : (
              <Button 
                size="sm"
                variant="outline"
                className="text-green-600 border-green-600"
                disabled
              >
                ✓ Added to Library
              </Button>
            )}
            {!searchStarted && (
              <Button 
                size="sm"
                variant="secondary"
                onClick={() => setSearchStarted(true)}
              >
                🔍 Manual Search
              </Button>
            )}
          </div>
          {addError && (
            <p className="text-sm text-destructive mt-2">{addError}</p>
          )}
        </div>
      </div>

      {/* Initial state - search not started */}
      {!searchStarted && (
        <Card>
          <CardContent className="py-12 text-center">
            <div className="text-4xl mb-4">🏴</div>
            <h3 className="text-lg font-semibold mb-2">Ready for manual search?</h3>
            <p className="text-muted-foreground mb-6 max-w-md mx-auto">
              Click the button below to search digital libraries and torrent indexers for "{book.title}"
            </p>
            <Button size="lg" onClick={() => setSearchStarted(true)}>
              🔍 Manual Search
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Loading Animation */}
      {searchStarted && isLoading && (
        <div className="py-6">
          <div className="relative h-24 mx-auto max-w-sm overflow-hidden rounded-lg">
            <div className="absolute bottom-4 left-2 text-2xl" title="Library">📚</div>
            <div className="absolute bottom-4 right-2 text-2xl" title="Your Collection">🌍</div>
            <div className="absolute bottom-3 w-16 h-12 animate-sail">
              <div className="animate-bob">
                <PirateShip className="w-full h-full drop-shadow-md" />
              </div>
            </div>
            <div className="absolute bottom-0 left-0 w-full overflow-hidden animate-waves">
              <WavesSVG />
            </div>
          </div>
          <div className="text-center mt-4">
            {sources.some(s => s.retrying) ? (
              <p className="text-sm text-amber-500 animate-pulse">Some sources had issues, retrying...</p>
            ) : (
              <p className="text-sm text-muted-foreground animate-pulse">Sailing the seven seas for books...</p>
            )}
          </div>
          <div className="flex justify-center gap-3 mt-3 flex-wrap">
            {sources.map((source) => (
              <div key={source.name} className="flex items-center gap-1.5 text-xs text-muted-foreground">
                {source.done ? (
                  source.error ? <span className="text-destructive">✗</span> : <span className="text-green-500">✓</span>
                ) : (
                  <div className="relative h-3.5 w-3.5">
                    <div className="absolute inset-0 rounded-full animate-gradient-spin" style={{
                      background: source.retrying
                        ? "conic-gradient(from 0deg, transparent, #f59e0b, #eab308, transparent)"
                        : "conic-gradient(from 0deg, transparent, #60a5fa, #3b82f6, transparent)",
                    }} />
                    <div className="absolute inset-0.5 rounded-full bg-background" />
                  </div>
                )}
                <span className="hidden sm:inline">
                  {source.name}
                  {source.retrying && <span className="text-amber-500 ml-1">(retrying...)</span>}
                </span>
              </div>
            ))}
          </div>

          <style>{`
            @keyframes sail { 0%, 100% { left: 10%; } 50% { left: calc(90% - 4rem); } }
            @keyframes bob { 0%, 100% { transform: translateY(0) rotate(-2deg); } 25% { transform: translateY(-3px) rotate(0deg); } 50% { transform: translateY(0) rotate(2deg); } 75% { transform: translateY(2px) rotate(0deg); } }
            @keyframes waves { 0% { transform: translateX(0); } 100% { transform: translateX(-50%); } }
            @keyframes gradient-spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
            .animate-sail { animation: sail 4s ease-in-out infinite; }
            .animate-bob { animation: bob 2s ease-in-out infinite; }
            .animate-waves { animation: waves 3s linear infinite; }
            .animate-gradient-spin { animation: gradient-spin 1s linear infinite; }
          `}</style>
        </div>
      )}

      {/* Filters */}
      {searchStarted && !isLoading && (libraryResults.length > 0 || indexerResults.length > 0) && (
        <div className="p-4 rounded-lg border bg-card space-y-4">
          <h3 className="font-semibold">Filters</h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <label className="text-sm text-muted-foreground block mb-1">Search in results</label>
              <Input
                placeholder="Filter results..."
                value={searchInResults}
                onChange={(e) => setSearchInResults(e.target.value)}
              />
            </div>
            <div>
              <label className="text-sm text-muted-foreground block mb-1">Format</label>
              <select
                value={formatFilter}
                onChange={(e) => setFormatFilter(e.target.value)}
                className="w-full h-9 px-3 rounded-md border bg-background text-sm"
              >
                {FORMAT_OPTIONS.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm text-muted-foreground block mb-1">Source</label>
              <select
                value={providerFilter}
                onChange={(e) => setProviderFilter(e.target.value)}
                className="w-full h-9 px-3 rounded-md border bg-background text-sm"
              >
                {PROVIDER_OPTIONS.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm text-muted-foreground block mb-1">Sort by</label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
                className="w-full h-9 px-3 rounded-md border bg-background text-sm"
              >
                {SORT_OPTIONS.map(opt => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
          </div>
        </div>
      )}

      {/* Warning for partial failures */}
      {searchStarted && !isLoading && someSourcesFailed && (
        <div className="bg-amber-500/10 border border-amber-500/30 rounded p-3 text-sm">
          <p className="text-amber-600 dark:text-amber-400 font-medium flex items-center gap-2">
            <span>⚠️</span> Some sources unavailable after retrying
          </p>
          <p className="text-amber-600/80 dark:text-amber-400/80 mt-1">
            {indexerFailed && "Torrent indexers (Prowlarr) could not be reached. "}
            {libraryFailed && "Digital libraries did not respond. "}
            Showing results from available sources.
          </p>
        </div>
      )}

      {/* Results */}
      {searchStarted && !isLoading && (
        <div className="space-y-6">
          {/* Digital Library Results */}
          {filteredLibraryResults.length > 0 && (
            <div>
              <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
                <span>📚</span> Digital Libraries
                <Badge variant="secondary">{filteredLibraryResults.length}</Badge>
              </h3>
              <div className="grid gap-2">
                {filteredLibraryResults.map((result) => (
                  <LibraryResult key={`${result.provider}-${result.id}`} result={result} />
                ))}
              </div>
            </div>
          )}

          {/* Indexer Results */}
          {filteredIndexerResults.length > 0 && (
            <div>
              <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
                <span>🔍</span> Torrent/Usenet Indexers
                <Badge variant="secondary">{filteredIndexerResults.length}</Badge>
              </h3>
              <div className="grid gap-2">
                {filteredIndexerResults.map((result) => (
                  <DownloadResult key={result.guid} result={result} />
                ))}
              </div>
            </div>
          )}

          {/* No results after filtering */}
          {totalResults === 0 && (libraryResults.length > 0 || indexerResults.length > 0) && (
            <div className="text-center text-muted-foreground py-8">
              <p className="text-lg mb-2">No results match your filters</p>
              <Button variant="outline" onClick={() => {
                setFormatFilter("all")
                setProviderFilter("all")
                setSearchInResults("")
              }}>
                Reset Filters
              </Button>
            </div>
          )}

          {/* No results at all */}
          {totalResults === 0 && libraryResults.length === 0 && indexerResults.length === 0 && (
            <div className="text-center text-muted-foreground py-8">
              <p className="text-lg mb-2">No downloads found</p>
              <p className="text-sm mb-4">
                Could not find "{book.title}" in any available source.
              </p>
            </div>
          )}

          {/* Footer */}
          {totalResults > 0 && (
            <div className="text-sm text-muted-foreground text-center pt-4 border-t">
              {totalResults} download {totalResults === 1 ? "option" : "options"} found
            </div>
          )}
        </div>
      )}
    </div>
  )
}
