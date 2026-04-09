import { useState, useMemo, useEffect } from "react"
import { useNavigate, Link } from "@tanstack/react-router"
import { useDigitalLibrarySearch } from "../../hooks/useMetadata"
import { useSearch, type SearchParams } from "../../hooks/useIndexers"
import { useCreateBook } from "../../hooks/useBooks"
import { useCreateAuthor, useAuthors } from "../../hooks/useAuthors"
import { useManualGrab } from "../../hooks/useWanted"
import { useRootFolders } from "../../hooks/useRootFolders"
import { Button, Card, CardContent, Badge, Input } from "../ui"
import type { MetadataBookResult } from "../../lib/api"
import { DownloadResult, LibraryResult } from "./SearchResultCards"
import { SearchLoadingAnimation } from "./SearchLoadingAnimation"

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

interface BookDetailsProps {
  book: MetadataBookResult
  autoSearch?: boolean
  existingBookId?: number
}

export function BookDetails({ book, autoSearch = false, existingBookId }: BookDetailsProps) {
  const navigate = useNavigate()
  
  // Search state - start automatically if autoSearch is true
  const [searchStarted, setSearchStarted] = useState(autoSearch)
  
  // Auto-start search when autoSearch prop changes
  useEffect(() => {
    if (autoSearch && !searchStarted) {
      setSearchStarted(true)
    }
  }, [autoSearch, searchStarted])
  
  // Add to library state
  const [addedToLibrary, setAddedToLibrary] = useState(!!existingBookId)
  const [addingToLibrary, setAddingToLibrary] = useState(false)
  const [addError, setAddError] = useState<string | null>(null)
  const [createdBookId, setCreatedBookId] = useState<number | null>(null)
  
  // Grab state
  const [isGrabbing, setIsGrabbing] = useState(false)
  const [grabSuccess, setGrabSuccess] = useState(false)
  const [grabError, setGrabError] = useState<string | null>(null)
  
  const createBook = useCreateBook()
  const createAuthor = useCreateAuthor()
  const manualGrab = useManualGrab()
  const { data: rootFolders } = useRootFolders()
  const authorName = book.authors?.[0] || "Unknown Author"
  const { data: existingAuthors } = useAuthors({ search: authorName, limit: 1 })
  
  const hasRootFolder = rootFolders && rootFolders.length > 0
  
  // Filters state
  const [formatFilter, setFormatFilter] = useState<string>("all")
  const [providerFilter, setProviderFilter] = useState<string>("all")
  const [sortBy, setSortBy] = useState<string>("score")
  const [searchInResults, setSearchInResults] = useState("")

  // Track if we have an existing book ID (for manual search from book page)
  const [bookIdToUse, setBookIdToUse] = useState<number | undefined>(existingBookId)

  // Ensure book is in library and return bookId
  const ensureBookInLibrary = async (): Promise<number> => {
    // If we already have a book ID (from existing book in library), use it
    if (bookIdToUse) {
      return bookIdToUse
    }
    
    if (createdBookId) {
      return createdBookId
    }
    
    setAddingToLibrary(true)
    setAddError(null)
    
    try {
      let authorId: number
      
      // Check if author already exists
      if (existingAuthors?.records?.length && existingAuthors.records[0].name.toLowerCase() === authorName.toLowerCase()) {
        authorId = existingAuthors.records[0].id
      } else {
        // Create author (service handles duplicates gracefully)
        try {
          const author = await createAuthor.mutateAsync({
            name: authorName,
            monitored: true,
          })
          authorId = author.id
        } catch (authorErr) {
          // If author creation fails due to conflict, search for existing
          if (authorErr instanceof Error && authorErr.message.includes("already exists")) {
            // Refetch authors and use the first match
            const response = await fetch(`/api/v1/authors?search=${encodeURIComponent(authorName)}&limit=1`)
            const data = await response.json()
            if (data.records?.length > 0) {
              authorId = data.records[0].id
            } else {
              throw authorErr
            }
          } else {
            throw authorErr
          }
        }
      }
      
      // Create book
      const newBook = await createBook.mutateAsync({
        authorId,
        title: book.title,
        foreignId: book.foreignId || "",
        isbn13: book.isbn13 || "",
        releaseDate: book.publishedYear ? `${book.publishedYear}-01-01` : "",
        imageUrl: book.coverUrl || "",
        monitored: true,
      })
      
      setCreatedBookId(newBook.id)
      setBookIdToUse(newBook.id)
      setAddedToLibrary(true)
      return newBook.id
    } catch (err) {
      setAddError(err instanceof Error ? err.message : "Failed to add to library")
      throw err
    } finally {
      setAddingToLibrary(false)
    }
  }

  // Handle add to library button
  const handleAddToLibrary = async () => {
    try {
      await ensureBookInLibrary()
    } catch {
      // Error already set in ensureBookInLibrary
    }
  }
  
  // Handle grab - ensures book is in library first, then sends to download
  const handleGrab = async (downloadUrl: string, releaseTitle: string, size: number) => {
    setIsGrabbing(true)
    setGrabError(null)
    setGrabSuccess(false)
    
    try {
      // First ensure the book is in the library
      const bookId = await ensureBookInLibrary()
      
      // Now send the grab request
      await manualGrab.mutateAsync({
        bookId,
        downloadUrl,
        releaseTitle,
        size,
      })
      
      setGrabSuccess(true)
    } catch (err) {
      setGrabError(err instanceof Error ? err.message : "Failed to grab release")
    } finally {
      setIsGrabbing(false)
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
          {grabError && (
            <p className="text-sm text-destructive mt-2">{grabError}</p>
          )}
        </div>
      </div>

      {/* Grab success notification */}
      {grabSuccess && (
        <Card className="border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950">
          <CardContent className="p-4">
            <p className="text-green-700 dark:text-green-300 flex items-center gap-2">
              <span>✓</span> Release grabbed! Check the Activity tab.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Initial state - search not started */}
      {!searchStarted && (
        <Card>
          <CardContent className="py-8">
            <div className="grid md:grid-cols-2 gap-8">
              {/* Wanted Section */}
              <div className="text-center md:text-left space-y-3">
                <div className="flex items-center gap-2 justify-center md:justify-start">
                  <span className="text-2xl">📋</span>
                  <h3 className="text-lg font-semibold">Add to Wanted</h3>
                </div>
                <p className="text-sm text-muted-foreground">
                  Click <strong>"Add to Library"</strong> above to monitor this book. 
                  Bookaneer will automatically search for it when new releases become available.
                </p>
                <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/50 p-2 rounded">
                  <span>💡</span>
                  <span>Wanted books appear in <strong>Activity → Wanted</strong> tab</span>
                </div>
              </div>

              {/* Manual Search Section */}
              <div className="text-center md:text-left space-y-3 md:border-l md:pl-8">
                <div className="flex items-center gap-2 justify-center md:justify-start">
                  <span className="text-2xl">🔍</span>
                  <h3 className="text-lg font-semibold">Manual Search</h3>
                </div>
                <p className="text-sm text-muted-foreground">
                  Search digital libraries and torrent indexers right now for "{book.title}".
                  Review results and grab what you want.
                </p>
                {!hasRootFolder && (
                  <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-md p-3">
                    <p className="text-sm text-yellow-600 dark:text-yellow-400 font-medium">⚠️ No Root Folder</p>
                    <p className="text-xs text-muted-foreground mt-1">
                      Configure a root folder in <Link to="/settings" className="text-primary hover:underline">Settings</Link> before downloading.
                    </p>
                  </div>
                )}
                <Button size="lg" onClick={() => setSearchStarted(true)} className="w-full md:w-auto">
                  🏴 Start Manual Search
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Loading Animation */}
      {searchStarted && isLoading && (
        <SearchLoadingAnimation sources={sources} />
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
                  <LibraryResult 
                    key={`${result.provider}-${result.id}`} 
                    result={result} 
                    onGrab={handleGrab}
                    isGrabbing={isGrabbing}
                  />
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
                  <DownloadResult 
                    key={result.guid} 
                    result={result} 
                    onGrab={handleGrab}
                    isGrabbing={isGrabbing}
                  />
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
