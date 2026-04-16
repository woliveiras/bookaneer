import { useState } from "react"
import { BookHeader } from "../../components/search/BookHeader"
import { SearchModal } from "../../components/search/SearchModal"
import { Card, CardContent } from "../../components/ui"
import { useBookRelease } from "../../hooks/useBookRelease"
import { useMetadataBook } from "../../hooks/useMetadata"
import { useRootFolders } from "../../hooks/useRootFolders"
import type { MetadataBookResult } from "../../lib/api"

interface BookDetailsProps {
  book: MetadataBookResult
  autoSearch?: boolean
  existingBookId?: number
}

export function BookDetails({ book, autoSearch = false, existingBookId }: BookDetailsProps) {
  const release = useBookRelease(book, existingBookId)
  const { data: rootFolders } = useRootFolders()
  const { data: bookMetadata } = useMetadataBook(book.foreignId, book.provider, !!book.foreignId)

  // Auto-start search when autoSearch prop is true
  const [autoStarted, setAutoStarted] = useState(false)
  if (autoSearch && !autoStarted && !release.searchStarted) {
    setAutoStarted(true)
    release.startSearch()
  }

  return (
    <div className="space-y-6">
      <BookHeader
        book={book}
        bookMetadata={bookMetadata}
        addedToLibrary={false}
        addingToLibrary={release.addingToLibrary}
        addError={release.addError}
        grabError={release.grabError}
        searchStarted={release.searchStarted}
        hasRootFolder={!!(rootFolders?.length)}
        onAddToLibrary={() => undefined}
        onStartSearch={release.startSearch}
      />

      {release.grabSuccess && (
        <Card className="border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950">
          <CardContent className="p-4">
            <p className="text-green-700 dark:text-green-300 flex items-center gap-2">
              <span>✓</span> Release grabbed! Check the Activity tab.
            </p>
          </CardContent>
        </Card>
      )}

      <SearchModal
        open={release.searchStarted}
        onClose={release.closeSearch}
        bookTitle={book.title}
        filteredLibraryResults={release.filteredLibraryResults}
        filteredIndexerResults={release.filteredIndexerResults}
        totalResults={release.totalResults}
        rawLibraryCount={release.rawLibraryCount}
        rawIndexerCount={release.rawIndexerCount}
        isLibraryLoading={release.isLibraryLoading}
        isIndexerLoading={release.isIndexerLoading}
        libraryFailed={release.libraryFailed}
        indexerFailed={release.indexerFailed}
        someSourcesFailed={release.someSourcesFailed}
        isGrabbing={release.isGrabbing}
        grabSuccess={release.grabSuccess}
        grabError={release.grabError}
        onGrab={release.handleGrab}
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
        onResetFilters={release.resetFilters}
        onExpandSearch={release.handleExpandSearch}
        isExpanded={false}
        isExpandSearching={release.isExpandSearching}
        libraryColumnConfig={release.libraryColumnConfig}
        indexerColumnConfig={release.indexerColumnConfig}
      />
    </div>
  )
}
