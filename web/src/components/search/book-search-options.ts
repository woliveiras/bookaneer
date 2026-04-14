// Format filter options
export const FORMAT_OPTIONS = [
  { value: "all", label: "All Formats" },
  { value: "epub", label: "EPUB" },
  { value: "pdf", label: "PDF" },
  { value: "mobi", label: "MOBI" },
  { value: "azw3", label: "AZW3" },
  { value: "cbz", label: "CBZ" },
  { value: "cbr", label: "CBR" },
] as const

// Language filter options
export const LANGUAGE_OPTIONS = [
  { value: "all", label: "All Languages" },
  { value: "en", label: "English" },
  { value: "pt", label: "Portuguese" },
  { value: "es", label: "Spanish" },
  { value: "fr", label: "French" },
  { value: "de", label: "German" },
  { value: "it", label: "Italian" },
  { value: "ru", label: "Russian" },
  { value: "zh", label: "Chinese" },
  { value: "ja", label: "Japanese" },
  { value: "ko", label: "Korean" },
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
  { value: "size-asc", label: "Smallest First" },
  { value: "size-desc", label: "Largest First" },
] as const
