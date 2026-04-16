import { useMemo, useState } from "react"
import type { MetadataBookResult } from "../lib/api"
import type { GrabResult } from "../lib/types"
import { useAuthors, useCreateAuthor } from "./useAuthors"
import { useCreateBook } from "./useBooks"
import { type SearchParams, useSearch } from "./useIndexers"
import { useDigitalLibrarySearch } from "./useMetadata"
import { useRootFolders } from "./useRootFolders"
import { useIndexerGrab, useManualGrab } from "./useWanted"

export interface IndexerGrabMeta {
  sourceType: "indexer"
  guid?: string
  seeders?: number
  indexerId?: number
  indexerName?: string
}

export interface LibraryGrabMeta {
  sourceType: "library"
}

export type GrabMeta = IndexerGrabMeta | LibraryGrabMeta

export function useBookRelease(book: MetadataBookResult | null, existingBookId?: number) {
  const [searchStarted, setSearchStarted] = useState(false)
  const [expandedSearch, setExpandedSearch] = useState(false)

  const [bookIdToUse, setBookIdToUse] = useState<number | undefined>(existingBookId)
  const [createdBookId, setCreatedBookId] = useState<number | null>(null)
  const [addingToLibrary, setAddingToLibrary] = useState(false)
  const [addError, setAddError] = useState<string | null>(null)

  const [isGrabbing, setIsGrabbing] = useState(false)
  const [grabSuccess, setGrabSuccess] = useState(false)
  const [grabError, setGrabError] = useState<string | null>(null)
  const [grabResult, setGrabResult] = useState<GrabResult | null>(null)

  const [formatFilter, setFormatFilter] = useState<string>("all")
  const [providerFilter, setProviderFilter] = useState<string>("all")
  const [languageFilter, setLanguageFilter] = useState<string>("all")
  const [sortBy, setSortBy] = useState<string>("score")
  const [searchInResults, setSearchInResults] = useState("")

  const createBook = useCreateBook()
  const createAuthor = useCreateAuthor()
  const manualGrab = useManualGrab()
  const indexerGrab = useIndexerGrab()
  const { data: rootFolders } = useRootFolders()
  const authorName = book?.authors?.[0] ?? "Unknown Author"
  const { data: existingAuthors } = useAuthors({ search: authorName, limit: 1 })

  const isbn = book?.isbn13 ?? ""
  // Use title+author for library search — Anna's Archive and LibGen return much
  // better results with a text query than a raw ISBN.
  const titleAuthorQuery = [book?.title ?? "", ...(book?.authors ?? [])].join(" ").trim()
  const librarySearchQuery = titleAuthorQuery || isbn
  const searchParams: SearchParams = isbn ? { isbn, q: titleAuthorQuery } : { q: titleAuthorQuery }

  const indexerSearch = useSearch(searchStarted ? searchParams : { q: "" }, searchStarted && !!book)
  const librarySearch = useDigitalLibrarySearch(librarySearchQuery, searchStarted && !!book)

  const expandedIndexerQuery = titleAuthorQuery
  const expandedIndexerSearch = useSearch(
    expandedSearch ? { q: expandedIndexerQuery } : { q: "" },
    expandedSearch && !!book,
  )
  const expandedLibrarySearch = useDigitalLibrarySearch(
    book?.title ?? "",
    expandedSearch && !!isbn && !!book,
  )

  const isExpandSearching =
    expandedSearch && (expandedIndexerSearch.isLoading || expandedLibrarySearch.isLoading)

  const libraryFailed =
    !librarySearch.isLoading && !!librarySearch.error && !librarySearch.data?.results?.length
  const indexerFailed =
    !indexerSearch.isLoading && !!indexerSearch.error && !indexerSearch.data?.results?.length
  const someSourcesFailed = (libraryFailed || indexerFailed) && !(libraryFailed && indexerFailed)

  // Track which result IDs came exclusively from the expanded search
  const expandedIndexerGuids = useMemo(() => {
    const primaryGuids = new Set((indexerSearch.data?.results ?? []).map((r) => r.guid))
    const set = new Set<string>()
    for (const r of expandedIndexerSearch.data?.results ?? []) {
      if (!primaryGuids.has(r.guid)) set.add(r.guid)
    }
    return set
  }, [indexerSearch.data, expandedIndexerSearch.data])

  const expandedLibraryKeys = useMemo(() => {
    const primaryKeys = new Set(
      (librarySearch.data?.results ?? []).map((r) => `${r.provider}-${r.id}`),
    )
    const set = new Set<string>()
    for (const r of expandedLibrarySearch.data?.results ?? []) {
      const key = `${r.provider}-${r.id}`
      if (!primaryKeys.has(key)) set.add(key)
    }
    return set
  }, [librarySearch.data, expandedLibrarySearch.data])

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
      const category = result.category?.toLowerCase() ?? ""
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

  const filteredLibraryResults = useMemo(() => {
    let results = [...libraryResults]

    if (formatFilter !== "all") {
      results = results.filter((r) => r.format.toLowerCase() === formatFilter)
    }
    if (languageFilter !== "all") {
      results = results.filter((r) => r.language?.toLowerCase().startsWith(languageFilter))
    }
    if (providerFilter !== "all" && providerFilter !== "torrent") {
      results = results.filter((r) => r.provider === providerFilter)
    }
    if (searchInResults.trim()) {
      const s = searchInResults.toLowerCase()
      results = results.filter(
        (r) =>
          r.title.toLowerCase().includes(s) || r.authors?.some((a) => a.toLowerCase().includes(s)),
      )
    }

    switch (sortBy) {
      case "year-desc":
        results.sort((a, b) => (b.year || 0) - (a.year || 0))
        break
      case "year-asc":
        results.sort((a, b) => (a.year || 0) - (b.year || 0))
        break
      case "size-asc":
        results.sort((a, b) => (a.size || 0) - (b.size || 0))
        break
      case "size-desc":
        results.sort((a, b) => (b.size || 0) - (a.size || 0))
        break
      case "format": {
        const order = { epub: 1, pdf: 2, mobi: 3 } as Record<string, number>
        results.sort(
          (a, b) => (order[a.format.toLowerCase()] ?? 99) - (order[b.format.toLowerCase()] ?? 99),
        )
        break
      }
      default:
        results.sort((a, b) => (b.score || 0) - (a.score || 0))
    }

    return results
  }, [libraryResults, formatFilter, languageFilter, providerFilter, sortBy, searchInResults])

  const filteredIndexerResults = useMemo(() => {
    if (providerFilter !== "all" && providerFilter !== "torrent") return []
    let results = [...indexerResults]
    if (searchInResults.trim()) {
      const s = searchInResults.toLowerCase()
      results = results.filter((r) => r.title.toLowerCase().includes(s))
    }
    return results
  }, [indexerResults, providerFilter, searchInResults])

  const totalResults = filteredLibraryResults.length + filteredIndexerResults.length

  const ensureBookInLibrary = async (): Promise<number> => {
    if (bookIdToUse) return bookIdToUse
    if (createdBookId) return createdBookId
    if (!book) throw new Error("No book selected")

    setAddingToLibrary(true)
    setAddError(null)

    try {
      let authorId: number

      if (
        existingAuthors?.records?.length &&
        existingAuthors.records[0].name.toLowerCase() === authorName.toLowerCase()
      ) {
        authorId = existingAuthors.records[0].id
      } else {
        try {
          const author = await createAuthor.mutateAsync({ name: authorName, monitored: true })
          authorId = author.id
        } catch (authorErr) {
          if (authorErr instanceof Error && authorErr.message.includes("already exists")) {
            const response = await fetch(
              `/api/v1/authors?search=${encodeURIComponent(authorName)}&limit=1`,
            )
            const data = (await response.json()) as { records?: Array<{ id: number }> }
            if (data.records?.length) {
              authorId = data.records[0].id
            } else {
              throw authorErr
            }
          } else {
            throw authorErr
          }
        }
      }

      const newBook = await createBook.mutateAsync({
        authorId,
        title: book.title,
        foreignId: book.foreignId ?? "",
        isbn13: book.isbn13 ?? "",
        releaseDate: book.publishedYear ? `${book.publishedYear}-01-01` : "",
        imageUrl: book.coverUrl ?? "",
        monitored: true,
      })

      setCreatedBookId(newBook.id)
      setBookIdToUse(newBook.id)
      return newBook.id
    } catch (err) {
      setAddError(err instanceof Error ? err.message : "Failed to add to library")
      throw err
    } finally {
      setAddingToLibrary(false)
    }
  }

  const handleGrab = async (
    downloadUrl: string,
    releaseTitle: string,
    size: number,
    meta?: GrabMeta,
  ) => {
    setIsGrabbing(true)
    setGrabError(null)
    setGrabSuccess(false)
    setGrabResult(null)

    try {
      const bookId = await ensureBookInLibrary()
      let result: GrabResult

      if (meta?.sourceType === "indexer") {
        result = await indexerGrab.mutateAsync({
          bookId,
          downloadUrl,
          releaseTitle,
          size,
          guid: meta.guid,
          seeders: meta.seeders,
          indexerId: meta.indexerId,
          indexerName: meta.indexerName,
        })
      } else {
        result = await manualGrab.mutateAsync({ bookId, downloadUrl, releaseTitle, size })
      }

      setGrabResult(result)
      setGrabSuccess(true)
    } catch (err) {
      setGrabError(err instanceof Error ? err.message : "Failed to grab release")
    } finally {
      setIsGrabbing(false)
    }
  }

  const startSearch = () => {
    setSearchStarted(true)
    setExpandedSearch(false)
    setGrabSuccess(false)
    setGrabError(null)
    setFormatFilter("all")
    setProviderFilter("all")
    setLanguageFilter("all")
    setSortBy("score")
    setSearchInResults("")
  }

  const closeSearch = () => {
    setSearchStarted(false)
    setExpandedSearch(false)
    setGrabSuccess(false)
    setGrabError(null)
    setFormatFilter("all")
    setProviderFilter("all")
    setLanguageFilter("all")
    setSortBy("score")
    setSearchInResults("")
  }

  return {
    searchStarted,
    startSearch,
    closeSearch,
    hasRootFolder: !!rootFolders?.length,
    addingToLibrary,
    addError,
    isExpandSearching,
    // Results
    filteredLibraryResults,
    filteredIndexerResults,
    totalResults,
    rawLibraryCount: libraryResults.length,
    rawIndexerCount: indexerResults.length,
    isLibraryLoading: librarySearch.isLoading,
    isIndexerLoading: indexerSearch.isLoading,
    libraryFailed,
    indexerFailed,
    someSourcesFailed,
    libraryColumnConfig: librarySearch.data?.columnConfig,
    indexerColumnConfig: indexerSearch.data?.columnConfig,
    // Grab
    isGrabbing,
    grabSuccess,
    grabError,
    grabResult,
    handleGrab,
    expandedLibraryKeys,
    expandedIndexerGuids,
    handleExpandSearch: () => setExpandedSearch(true),
    isExpanded: expandedSearch,
    // Filters
    formatFilter,
    setFormatFilter,
    providerFilter,
    setProviderFilter,
    languageFilter,
    setLanguageFilter,
    sortBy,
    setSortBy,
    searchInResults,
    setSearchInResults,
    resetFilters: () => {
      setFormatFilter("all")
      setLanguageFilter("all")
      setProviderFilter("all")
      setSearchInResults("")
    },
  }
}
