import { useQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { AlertTriangle, ArrowLeft, Library } from "lucide-react"
import { useEffect, useState } from "react"
import { toast } from "sonner"
import { AuthLayout } from "../components/layout/AppLayout"
import { SearchFilters } from "../components/search/SearchFilters"
import { SearchResults } from "../components/search/SearchResults"
import { useBookRelease } from "../hooks/useBookRelease"
import type { MetadataBookResult } from "../lib/api"
import { bookApi } from "../lib/api"

interface ReleasesPageProps {
  book: MetadataBookResult
  autoSearch?: boolean
  existingBookId?: number
}

export function ReleasesPage({ book, autoSearch = true, existingBookId }: ReleasesPageProps) {
  const release = useBookRelease(book, existingBookId)

  const { data: existingBook } = useQuery({
    queryKey: ["book", existingBookId],
    queryFn: () => bookApi.get(existingBookId!),
    enabled: !!existingBookId,
  })
  const hasExistingFile = !!(existingBook?.files && existingBook.files.length > 0)

  // Auto-start search on first render
  const [autoStarted, setAutoStarted] = useState(false)
  if (autoSearch && !autoStarted && !release.searchStarted) {
    setAutoStarted(true)
    release.startSearch()
  }

  useEffect(() => {
    if (release.grabSuccess) {
      toast.success(
        release.grabResult?.clientName
          ? `Sent to download!`
          : "Release grabbed! Check the Activity tab.",
      )
    }
  }, [release.grabSuccess, release.grabResult])

  useEffect(() => {
    if (release.grabError) {
      toast.error(release.grabError)
    }
  }, [release.grabError])

  return (
    <AuthLayout>
      <div className="space-y-6">
        {/* Back + Book header */}
        <div className="flex items-start gap-4">
          <Link
            to="/search"
            className="mt-1 p-2 rounded-md hover:bg-muted transition-colors shrink-0"
            aria-label="Back to search"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>

          <div className="flex gap-4 flex-1 min-w-0">
            {book.coverUrl ? (
              <img
                src={book.coverUrl}
                alt={book.title}
                className="w-14 h-20 object-cover rounded shadow-sm shrink-0"
                loading="eager"
              />
            ) : (
              <div className="w-14 h-20 bg-muted rounded flex items-center justify-center shrink-0">
                <Library className="w-6 h-6 text-muted-foreground" />
              </div>
            )}
            <div className="min-w-0">
              <h1 className="text-xl font-bold leading-tight truncate">{book.title}</h1>
              {book.authors && book.authors.length > 0 && (
                <p className="text-sm text-muted-foreground mt-0.5">{book.authors.join(", ")}</p>
              )}
              {book.publishedYear && (
                <p className="text-xs text-muted-foreground">{book.publishedYear}</p>
              )}
              <p className="text-xs text-muted-foreground mt-1">
                {release.searchStarted
                  ? release.totalResults > 0
                    ? `${release.totalResults} releases found`
                    : release.isLibraryLoading || release.isIndexerLoading
                      ? "Searching…"
                      : "No releases found"
                  : "Releases"}
              </p>
            </div>
          </div>
        </div>

        {/* Partial failure warning */}
        {release.someSourcesFailed && (
          <div className="bg-amber-500/10 border border-amber-500/30 rounded p-3 text-sm">
            <p className="text-amber-600 dark:text-amber-400 font-medium flex items-center gap-2">
              <AlertTriangle className="w-4 h-4" /> Some sources unavailable
            </p>
            <p className="text-amber-600/80 dark:text-amber-400/80 mt-1">
              {release.indexerFailed && "Torrent indexers could not be reached. "}
              {release.libraryFailed && "Digital libraries did not respond. "}
              Showing results from available sources.
            </p>
          </div>
        )}

        {/* Filters — only after we have some results */}
        {(release.rawLibraryCount > 0 || release.rawIndexerCount > 0) && (
          <SearchFilters
            searchInResults={release.searchInResults}
            formatFilter={release.formatFilter}
            languageFilter={release.languageFilter}
            providerFilter={release.providerFilter}
            sortBy={release.sortBy}
            onSearchChange={release.setSearchInResults}
            onFormatChange={release.setFormatFilter}
            onLanguageChange={release.setLanguageFilter}
            onProviderChange={release.setProviderFilter}
            onSortChange={release.setSortBy}
          />
        )}

        {/* Results */}
        <SearchResults
          filteredLibraryResults={release.filteredLibraryResults}
          filteredIndexerResults={release.filteredIndexerResults}
          totalResults={release.totalResults}
          rawLibraryCount={release.rawLibraryCount}
          rawIndexerCount={release.rawIndexerCount}
          bookTitle={book.title}
          isGrabbing={release.isGrabbing}
          isLibraryLoading={release.isLibraryLoading}
          isIndexerLoading={release.isIndexerLoading}
          libraryError={release.libraryFailed}
          indexerError={release.indexerFailed}
          searchActive={release.searchStarted}
          onGrab={release.handleGrab}
          onResetFilters={release.resetFilters}
          onExpandSearch={release.handleExpandSearch}
          isExpanded={release.isExpanded}
          isExpandSearching={release.isExpandSearching}
          libraryColumnConfig={release.libraryColumnConfig}
          indexerColumnConfig={release.indexerColumnConfig}
          expandedLibraryKeys={release.expandedLibraryKeys}
          expandedIndexerGuids={release.expandedIndexerGuids}
          hasExistingFile={hasExistingFile}
        />
      </div>
    </AuthLayout>
  )
}
