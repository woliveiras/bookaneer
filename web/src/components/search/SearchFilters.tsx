import { useState } from "react"
import { ChevronDown, ChevronUp, SlidersHorizontal } from "lucide-react"
import { Input } from "../ui"
import { FORMAT_OPTIONS, LANGUAGE_OPTIONS, PROVIDER_OPTIONS, SORT_OPTIONS } from "./book-search-options"

interface SearchFiltersProps {
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
}

export function SearchFilters({
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
}: SearchFiltersProps) {
  const [open, setOpen] = useState(false)

  return (
    <div className="rounded-lg border bg-card">
      {/* Toggle button — always visible on mobile, hidden on sm+ */}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex sm:hidden items-center justify-between w-full px-4 py-3 text-sm font-semibold"
        aria-expanded={open}
      >
        <span className="flex items-center gap-2">
          <SlidersHorizontal className="w-4 h-4" aria-hidden="true" />
          Filters
        </span>
        {open ? (
          <ChevronUp className="w-4 h-4" aria-hidden="true" />
        ) : (
          <ChevronDown className="w-4 h-4" aria-hidden="true" />
        )}
      </button>

      {/* Content — always visible on sm+, toggle on mobile */}
      <div className={`p-4 space-y-4 ${open ? "block" : "hidden sm:block"}`}>
        <h3 className="font-semibold hidden sm:block">Filters</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 gap-4">
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
              Language
              <select
                value={languageFilter}
                onChange={(e) => onLanguageChange(e.target.value)}
                className="w-full h-9 px-3 rounded-md border bg-background text-sm"
              >
                {LANGUAGE_OPTIONS.map((opt) => (
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
    </div>
  )
}
