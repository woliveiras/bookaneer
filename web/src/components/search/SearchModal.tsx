import { AlertTriangle } from "lucide-react"
import type { DigitalLibraryResult } from "../../lib/api"
import type { ColumnConfig } from "../../lib/types"
import type { SearchResult } from "../../lib/types/indexer"
import { Dialog, DialogBody, DialogHeader } from "../ui"
import { SearchFilters } from "./SearchFilters"
import { SearchResults } from "./SearchResults"

interface SearchModalProps {
  open: boolean
  onClose: () => void
  bookTitle: string

  // Results
  filteredLibraryResults: DigitalLibraryResult[]
  filteredIndexerResults: SearchResult[]
  totalResults: number
  rawLibraryCount: number
  rawIndexerCount: number

  // Loading
  isLibraryLoading: boolean
  isIndexerLoading: boolean
  libraryFailed: boolean
  indexerFailed: boolean
  someSourcesFailed: boolean

  // Grab
  isGrabbing: boolean
  grabSuccess: boolean
  grabError: string | null
  onGrab: (downloadUrl: string, releaseTitle: string, size: number) => Promise<void>

  // Filters
  searchInResults: string
  formatFilter: string
  languageFilter: string
  providerFilter: string
  sortBy: string
  onSearchChange: (value: string) => void
  onFormatChange: (value: string) => void
  onLanguageChange: (value: string) => void
  onProviderChange: (value: string) => void
  onSortChange: (value: string) => void
  onResetFilters: () => void

  // Expand
  onExpandSearch: () => void
  isExpanded: boolean
  isExpandSearching: boolean

  // Column configs
  libraryColumnConfig?: ColumnConfig
  indexerColumnConfig?: ColumnConfig
}

export function SearchModal({
  open,
  onClose,
  bookTitle,
  filteredLibraryResults,
  filteredIndexerResults,
  totalResults,
  rawLibraryCount,
  rawIndexerCount,
  isLibraryLoading,
  isIndexerLoading,
  libraryFailed,
  indexerFailed,
  someSourcesFailed,
  isGrabbing,
  grabSuccess,
  grabError,
  onGrab,
  searchInResults,
  formatFilter,
  languageFilter,
  providerFilter,
  sortBy,
  onSearchChange,
  onFormatChange,
  onLanguageChange,
  onProviderChange,
  onSortChange,
  onResetFilters,
  onExpandSearch,
  isExpanded,
  isExpandSearching,
  libraryColumnConfig,
  indexerColumnConfig,
}: SearchModalProps) {
  return (
    <Dialog open={open} onClose={onClose}>
      <DialogHeader onClose={onClose}>
        <h2 className="text-lg font-semibold truncate">
          Search releases for "{bookTitle}"
        </h2>
      </DialogHeader>

      <DialogBody className="space-y-4">
        {/* Grab notifications */}
        {grabSuccess && (
          <div className="rounded-md border border-green-200 bg-green-50 dark:border-green-900 dark:bg-green-950 p-3">
            <p className="text-green-700 dark:text-green-300 text-sm flex items-center gap-2">
              <span>✓</span> Release grabbed! Check the Activity tab.
            </p>
          </div>
        )}
        {grabError && (
          <div className="rounded-md border border-destructive/30 bg-destructive/10 p-3">
            <p className="text-destructive text-sm">{grabError}</p>
          </div>
        )}

        {/* Filters */}
        {(rawLibraryCount > 0 || rawIndexerCount > 0) && (
          <SearchFilters
            searchInResults={searchInResults}
            formatFilter={formatFilter}
            languageFilter={languageFilter}
            providerFilter={providerFilter}
            sortBy={sortBy}
            onSearchChange={onSearchChange}
            onFormatChange={onFormatChange}
            onLanguageChange={onLanguageChange}
            onProviderChange={onProviderChange}
            onSortChange={onSortChange}
          />
        )}

        {/* Partial failure warning */}
        {someSourcesFailed && (
          <div className="bg-amber-500/10 border border-amber-500/30 rounded p-3 text-sm">
            <p className="text-amber-600 dark:text-amber-400 font-medium flex items-center gap-2">
              <span><AlertTriangle className="w-4 h-4" /></span> Some sources unavailable
            </p>
            <p className="text-amber-600/80 dark:text-amber-400/80 mt-1">
              {indexerFailed && "Torrent indexers could not be reached. "}
              {libraryFailed && "Digital libraries did not respond. "}
              Showing results from available sources.
            </p>
          </div>
        )}

        {/* Results with tabs */}
        <SearchResults
          filteredLibraryResults={filteredLibraryResults}
          filteredIndexerResults={filteredIndexerResults}
          totalResults={totalResults}
          rawLibraryCount={rawLibraryCount}
          rawIndexerCount={rawIndexerCount}
          bookTitle={bookTitle}
          isGrabbing={isGrabbing}
          isLibraryLoading={isLibraryLoading}
          isIndexerLoading={isIndexerLoading}
          libraryError={libraryFailed}
          indexerError={indexerFailed}
          searchActive
          onGrab={onGrab}
          onResetFilters={onResetFilters}
          onExpandSearch={onExpandSearch}
          isExpanded={isExpanded}
          isExpandSearching={isExpandSearching}
          libraryColumnConfig={libraryColumnConfig}
          indexerColumnConfig={indexerColumnConfig}
        />
      </DialogBody>
    </Dialog>
  )
}
