import type { DigitalLibraryResult, SearchResult } from "../../lib/api"
import type { ColumnConfig } from "../../lib/types"
import { Library, Search } from "lucide-react"
import { Badge, Button } from "../ui"
import { DownloadResult, LibraryResult } from "./SearchResultCards"

interface SearchResultsProps {
  filteredLibraryResults: DigitalLibraryResult[]
  filteredIndexerResults: SearchResult[]
  totalResults: number
  rawLibraryCount: number
  rawIndexerCount: number
  bookTitle: string
  isGrabbing: boolean
  onGrab: (downloadUrl: string, releaseTitle: string, size: number) => Promise<void>
  onResetFilters: () => void
  onExpandSearch?: () => void
  isExpanded?: boolean
  isExpandSearching?: boolean
  libraryColumnConfig?: ColumnConfig
  indexerColumnConfig?: ColumnConfig
}

export function SearchResults({
  filteredLibraryResults,
  filteredIndexerResults,
  totalResults,
  rawLibraryCount,
  rawIndexerCount,
  bookTitle,
  isGrabbing,
  onGrab,
  onResetFilters,
  onExpandSearch,
  isExpanded = false,
  isExpandSearching = false,
  libraryColumnConfig,
  indexerColumnConfig,
}: SearchResultsProps) {
  return (
    <div className="space-y-6">
      {/* Digital Library Results */}
      {filteredLibraryResults.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
            <span><Library className="w-5 h-5" /></span> Digital Libraries
            <Badge variant="secondary">{filteredLibraryResults.length}</Badge>
          </h3>
          <div className="grid gap-2">
            {filteredLibraryResults.map((result) => (
              <LibraryResult
                key={`${result.provider}-${result.id}`}
                result={result}
                onGrab={onGrab}
                isGrabbing={isGrabbing}
                columnConfig={libraryColumnConfig}
              />
            ))}
          </div>
        </div>
      )}

      {/* Indexer Results */}
      {filteredIndexerResults.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
            <span><Search className="w-5 h-5" /></span> Torrent/Usenet Indexers
            <Badge variant="secondary">{filteredIndexerResults.length}</Badge>
          </h3>
          <div className="grid gap-2">
            {filteredIndexerResults.map((result) => (
              <DownloadResult
                key={result.guid}
                result={result}
                onGrab={onGrab}
                isGrabbing={isGrabbing}
                columnConfig={indexerColumnConfig}
              />
            ))}
          </div>
        </div>
      )}

      {/* No results after filtering */}
      {totalResults === 0 && (rawLibraryCount > 0 || rawIndexerCount > 0) && (
        <div className="text-center text-muted-foreground py-8">
          <p className="text-lg mb-2">No results match your filters</p>
          <Button variant="outline" onClick={onResetFilters}>
            Reset Filters
          </Button>
        </div>
      )}

      {/* No results at all */}
      {totalResults === 0 && rawLibraryCount === 0 && rawIndexerCount === 0 && (
        <div className="text-center text-muted-foreground py-8">
          <p className="text-lg mb-2">No downloads found</p>
          <p className="text-sm mb-4">Could not find "{bookTitle}" in any available source.</p>
          {onExpandSearch && !isExpanded && (
            <Button variant="outline" onClick={onExpandSearch} disabled={isExpandSearching}>
              {isExpandSearching ? "Searching…" : "Expand search"}
            </Button>
          )}
        </div>
      )}

      {/* Footer */}
      {totalResults > 0 && (
        <div className="text-sm text-muted-foreground text-center pt-4 border-t space-y-2">
          <p>{totalResults} download {totalResults === 1 ? "option" : "options"} found</p>
          {onExpandSearch && !isExpanded && (
            <Button variant="ghost" size="sm" onClick={onExpandSearch} disabled={isExpandSearching}>
              {isExpandSearching ? "Searching…" : "Expand search"}
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
