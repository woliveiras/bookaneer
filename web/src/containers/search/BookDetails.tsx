import { useEffect, useMemo, useState } from "react"
import { AlertTriangle } from "lucide-react"
import { BookHeader } from "../../components/search/BookHeader"
import { SearchFilters } from "../../components/search/SearchFilters"
import { SearchLoadingAnimation } from "../../components/search/SearchLoadingAnimation"
import { SearchResults } from "../../components/search/SearchResults"
import { Card, CardContent } from "../../components/ui"
import { useAuthors, useCreateAuthor } from "../../hooks/useAuthors"
import { useCreateBook } from "../../hooks/useBooks"
import { type SearchParams, useSearch } from "../../hooks/useIndexers"
import { useDigitalLibrarySearch, useMetadataBook } from "../../hooks/useMetadata"
import { useRootFolders } from "../../hooks/useRootFolders"
import { useManualGrab } from "../../hooks/useWanted"
import type { MetadataBookResult } from "../../lib/api"

interface BookDetailsProps {
  book: MetadataBookResult
  autoSearch?: boolean
  existingBookId?: number
}

export function BookDetails({ book, autoSearch = false, existingBookId }: BookDetailsProps) {
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

  // Fetch full book metadata (description, genres, rating, links)
  const { data: bookMetadata } = useMetadataBook(book.foreignId, book.provider, !!book.foreignId)

  // Filters state
  const [formatFilter, setFormatFilter] = useState<string>("all")
  const [providerFilter, setProviderFilter] = useState<string>("all")
  const [sortBy, setSortBy] = useState<string>("score")
  const [searchInResults, setSearchInResults] = useState("")

  // Expand search state — broader query without ISBN/format restrictions
  const [expandedSearch, setExpandedSearch] = useState(false)

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
      if (
        existingAuthors?.records?.length &&
        existingAuthors.records[0].name.toLowerCase() === authorName.toLowerCase()
      ) {
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
            const response = await fetch(
              `/api/v1/authors?search=${encodeURIComponent(authorName)}&limit=1`,
            )
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

  // Build search queries — ISBN-first for more precise results
  const isbn = book.isbn13 || book.isbn10 || ""
  const librarySearchQuery = isbn || book.title
  const searchParams: SearchParams = isbn
    ? { isbn, q: [book.title, ...(book.authors || [])].join(" ") }
    : { q: [book.title, ...(book.authors || [])].join(" ") }

  // Search queries - only enabled when searchStarted is true
  const indexerSearch = useSearch(searchStarted ? searchParams : { q: "" }, searchStarted)
  const librarySearch = useDigitalLibrarySearch(librarySearchQuery, searchStarted)

  // Expanded search — uses just title (broader, no ISBN)
  const expandedLibraryQuery = book.title
  const expandedIndexerQuery = [book.title, ...(book.authors || [])].join(" ")
  const expandedIndexerSearch = useSearch(
    expandedSearch ? { q: expandedIndexerQuery } : { q: "" },
    expandedSearch,
  )
  const expandedLibrarySearch = useDigitalLibrarySearch(expandedLibraryQuery, expandedSearch && !!isbn)

  // Loading and error states
  const libraryDone = !librarySearch.isLoading
  const indexerDone = !indexerSearch.isLoading
  const isLoading = !libraryDone || !indexerDone
  const isExpandSearching = expandedSearch && (expandedIndexerSearch.isLoading || expandedLibrarySearch.isLoading)

  const libraryFailed = libraryDone && !!librarySearch.error && !librarySearch.data?.results?.length
  const indexerFailed = indexerDone && !!indexerSearch.error && !indexerSearch.data?.results?.length
  const someSourcesFailed = (libraryFailed || indexerFailed) && !(libraryFailed && indexerFailed)

  // Filter indexer results for ebooks — merge expanded results
  const indexerResults = useMemo(() => {
    const primary = indexerSearch.data?.results ?? []
    const expanded = expandedIndexerSearch.data?.results ?? []
    const seenGuids = new Set(primary.map((r) => r.guid))
    const merged = [...primary]
    for (const r of expanded) {
      if (!seenGuids.has(r.guid)) {
        seenGuids.add(r.guid)
        merged.push(r)
      }
    }
    return merged.filter((result) => {
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
  }, [indexerSearch.data, expandedIndexerSearch.data])

  // Merge library results with expanded library results (deduped by id+provider)
  const libraryResults = useMemo(() => {
    const primary = librarySearch.data?.results ?? []
    const expanded = expandedLibrarySearch.data?.results ?? []
    const seenKeys = new Set(primary.map((r) => `${r.provider}-${r.id}`))
    const merged = [...primary]
    for (const r of expanded) {
      const key = `${r.provider}-${r.id}`
      if (!seenKeys.has(key)) {
        seenKeys.add(key)
        merged.push(r)
      }
    }
    return merged
  }, [librarySearch.data, expandedLibrarySearch.data])

  // Apply filters and sorting
  const filteredLibraryResults = useMemo(() => {
    let results = [...libraryResults]

    // Format filter
    if (formatFilter !== "all") {
      results = results.filter((r) => r.format.toLowerCase() === formatFilter)
    }

    // Provider filter
    if (providerFilter !== "all" && providerFilter !== "torrent") {
      results = results.filter((r) => r.provider === providerFilter)
    }

    // Text search in results
    if (searchInResults.trim()) {
      const search = searchInResults.toLowerCase()
      results = results.filter(
        (r) =>
          r.title.toLowerCase().includes(search) ||
          r.authors?.some((a) => a.toLowerCase().includes(search)),
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
      case "format": {
        const formatOrder = { epub: 1, pdf: 2, mobi: 3 }
        results.sort((a, b) => {
          const aOrder = formatOrder[a.format.toLowerCase() as keyof typeof formatOrder] || 99
          const bOrder = formatOrder[b.format.toLowerCase() as keyof typeof formatOrder] || 99
          return aOrder - bOrder
        })
        break
      }
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
      results = results.filter((r) => r.title.toLowerCase().includes(search))
    }

    return results
  }, [indexerResults, providerFilter, searchInResults])

  const totalResults = filteredLibraryResults.length + filteredIndexerResults.length

  return (
    <div className="space-y-6">
      <BookHeader
        book={book}
        bookMetadata={bookMetadata}
        addedToLibrary={addedToLibrary}
        addingToLibrary={addingToLibrary}
        addError={addError}
        grabError={grabError}
        searchStarted={searchStarted}
        hasRootFolder={!!hasRootFolder}
        onAddToLibrary={handleAddToLibrary}
        onStartSearch={() => setSearchStarted(true)}
      />

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

      {/* Loading Animation */}
      {searchStarted && isLoading && <SearchLoadingAnimation />}

      {/* Filters */}
      {searchStarted && !isLoading && (libraryResults.length > 0 || indexerResults.length > 0) && (
        <SearchFilters
          searchInResults={searchInResults}
          formatFilter={formatFilter}
          providerFilter={providerFilter}
          sortBy={sortBy}
          onSearchChange={setSearchInResults}
          onFormatChange={setFormatFilter}
          onProviderChange={setProviderFilter}
          onSortChange={setSortBy}
        />
      )}

      {/* Warning for partial failures */}
      {searchStarted && !isLoading && someSourcesFailed && (
        <div className="bg-amber-500/10 border border-amber-500/30 rounded p-3 text-sm">
          <p className="text-amber-600 dark:text-amber-400 font-medium flex items-center gap-2">
            <span><AlertTriangle className="w-4 h-4" /></span> Some sources unavailable after retrying
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
        <SearchResults
          filteredLibraryResults={filteredLibraryResults}
          filteredIndexerResults={filteredIndexerResults}
          totalResults={totalResults}
          rawLibraryCount={libraryResults.length}
          rawIndexerCount={indexerResults.length}
          bookTitle={book.title}
          isGrabbing={isGrabbing}
          onGrab={handleGrab}
          onResetFilters={() => {
            setFormatFilter("all")
            setProviderFilter("all")
            setSearchInResults("")
          }}
          onExpandSearch={() => setExpandedSearch(true)}
          isExpanded={expandedSearch}
          isExpandSearching={isExpandSearching}
          libraryColumnConfig={librarySearch.data?.columnConfig}
          indexerColumnConfig={indexerSearch.data?.columnConfig}
        />
      )}
    </div>
  )
}
