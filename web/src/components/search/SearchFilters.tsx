import { Input } from "../ui"
import { FORMAT_OPTIONS, PROVIDER_OPTIONS, SORT_OPTIONS } from "./book-search-options"

interface SearchFiltersProps {
  searchInResults: string
  formatFilter: string
  providerFilter: string
  sortBy: string
  onSearchChange: (value: string) => void
  onFormatChange: (value: string) => void
  onProviderChange: (value: string) => void
  onSortChange: (value: string) => void
}

export function SearchFilters({
  searchInResults,
  formatFilter,
  providerFilter,
  sortBy,
  onSearchChange,
  onFormatChange,
  onProviderChange,
  onSortChange,
}: SearchFiltersProps) {
  return (
    <div className="p-4 rounded-lg border bg-card space-y-4">
      <h3 className="font-semibold">Filters</h3>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
        <div>
          {/* biome-ignore lint/a11y/noLabelWithoutControl: Input is a custom component wrapping native input */}
          <label className="text-sm text-muted-foreground block mb-1">
            Search in results
            <Input
              placeholder="Filter results..."
              value={searchInResults}
              onChange={(e) => onSearchChange(e.target.value)}
            />
          </label>
        </div>
        <div>
          <label className="text-sm text-muted-foreground block mb-1">
            Format
            <select
              value={formatFilter}
              onChange={(e) => onFormatChange(e.target.value)}
              className="w-full h-9 px-3 rounded-md border bg-background text-sm"
            >
              {FORMAT_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div>
          <label className="text-sm text-muted-foreground block mb-1">
            Source
            <select
              value={providerFilter}
              onChange={(e) => onProviderChange(e.target.value)}
              className="w-full h-9 px-3 rounded-md border bg-background text-sm"
            >
              {PROVIDER_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div>
          <label className="text-sm text-muted-foreground block mb-1">
            Sort by
            <select
              value={sortBy}
              onChange={(e) => onSortChange(e.target.value)}
              className="w-full h-9 px-3 rounded-md border bg-background text-sm"
            >
              {SORT_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </label>
        </div>
      </div>
    </div>
  )
}
