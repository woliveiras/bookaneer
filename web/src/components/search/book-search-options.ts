// Format filter options
export const FORMAT_OPTIONS = [
  { value: "all", label: "All Formats" },
  { value: "epub", label: "EPUB" },
  { value: "pdf", label: "PDF" },
  { value: "mobi", label: "MOBI" },
] as const

// Provider filter options
export const PROVIDER_OPTIONS = [
  { value: "all", label: "All Sources" },
  { value: "internet-archive", label: "Internet Archive" },
  { value: "libgen", label: "LibGen" },
  { value: "annas-archive", label: "Anna's Archive" },
  { value: "torrent", label: "Torrent Indexers" },
] as const

// Sort options
export const SORT_OPTIONS = [
  { value: "score", label: "Best Match" },
  { value: "year-desc", label: "Newest First" },
  { value: "year-asc", label: "Oldest First" },
  { value: "format", label: "By Format" },
] as const
