import { useState } from "react"
import type { DigitalLibraryResult, SearchResult } from "../../lib/api"
import type { ColumnConfig } from "../../lib/types"
import { Library, Search } from "lucide-react"
import { Badge, Button } from "../ui"
import { DownloadResult, LibraryResult } from "./SearchResultCards"

type SourceTab = "all" | "library" | "indexer"

interface SearchResultsProps {
  filteredLibraryResults: DigitalLibraryResult[]
  filteredIndexerResults: SearchResult[]
  totalResults: number
  rawLibraryCount: number
  rawIndexerCount: number
  bookTitle: string
  isGrabbing: boolean
  isLibraryLoading?: boolean
  isIndexerLoading?: boolean
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
  isLibraryLoading = false,
  isIndexerLoading = false,
  onGrab,
  onResetFilters,
  onExpandSearch,
  isExpanded = false,
  isExpandSearching = false,
  libraryColumnConfig,
  indexerColumnConfig,
}: SearchResultsProps) {
  const [activeTab, setActiveTab] = useState<SourceTab>("all")

  const showLibrary = activeTab === "all" || activeTab === "library"
  const showIndexer = activeTab === "all" || activeTab === "indexer"

  return (
    <div className="space-y-4">
      {/* Source tabs */}
      {(filteredLibraryResults.length > 0 || filteredIndexerResults.length > 0 || isLibraryLoading || isIndexerLoading) && (
        <div className="flex gap-1 border-b" role="tablist" aria-label="Result sources">
          <TabButton
            active={activeTab === "all"}
            onClick={() => setActiveTab("all")}
            label="All"
            count={totalResults}
          />
          <TabButton
            active={activeTab === "library"}
            onClick={() => setActiveTab("library")}
            icon={<Library className="w-4 h-4" />}
            label="Libraries"
            count={filteredLibraryResults.length}
            loading={isLibraryLoading}
          />
          <TabButton
            active={activeTab === "indexer"}
            onClick={() => setActiveTab("indexer")}
            icon={<Search className="w-4 h-4" />}
            label="Indexers"
            count={filteredIndexerResults.length}
            loading={isIndexerLoading}
          />
        </div>
      )}

      {/* Digital Library Results */}
      {showLibrary && filteredLibraryResults.length > 0 && (
        <div>
          {activeTab === "all" && (
            <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
              <span><Library className="w-5 h-5" /></span> Digital Libraries
              <Badge variant="secondary">{filteredLibraryResults.length}</Badge>
            </h3>
          )}
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

      {/* Library loading in tab view */}
      {showLibrary && isLibraryLoading && activeTab === "library" && (
        <p className="text-center text-muted-foreground py-6 text-sm">Searching libraries…</p>
      )}

      {/* Indexer Results */}
      {showIndexer && filteredIndexerResults.length > 0 && (
        <div>
          {activeTab === "all" && (
            <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
              <span><Search className="w-5 h-5" /></span> Torrent/Usenet Indexers
              <Badge variant="secondary">{filteredIndexerResults.length}</Badge>
            </h3>
          )}
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

      {/* Indexer loading in tab view */}
      {showIndexer && isIndexerLoading && activeTab === "indexer" && (
        <p className="text-center text-muted-foreground py-6 text-sm">Searching indexers…</p>
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
      {totalResults === 0 && rawLibraryCount === 0 && rawIndexerCount === 0 && !isLibraryLoading && !isIndexerLoading && (
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

interface TabButtonProps {
  active: boolean
  onClick: () => void
  icon?: React.ReactNode
  label: string
  count: number
  loading?: boolean
}

function TabButton({ active, onClick, icon, label, count, loading }: TabButtonProps) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      className={`flex items-center gap-1.5 px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
        active
          ? "border-primary text-primary"
          : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
      }`}
    >
      {icon}
      {label}
      {loading ? (
        <span className="inline-block w-3 h-3 border-2 border-muted-foreground border-t-transparent rounded-full animate-spin" />
      ) : (
        <Badge variant="secondary" className="text-xs ml-1">{count}</Badge>
      )}
    </button>
  )
}
