import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { Bookmark, ReaderBookFile, ReadingProgress } from "../lib/api"
import {
  useBookmarks,
  useCreateBookmark,
  useDeleteBookmark,
  useReaderBookFile,
  useReadingProgress,
  useSaveProgress,
} from "./useReader"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    readerApi: {
      getBookFile: vi.fn(),
      getProgress: vi.fn(),
      saveProgress: vi.fn(),
      listBookmarks: vi.fn(),
      createBookmark: vi.fn(),
      deleteBookmark: vi.fn(),
    },
  }
})

import { readerApi } from "../lib/api"

const mockBookFile: ReaderBookFile = {
  id: 1,
  bookId: 10,
  path: "/books/test.epub",
  relativePath: "test.epub",
  size: 500000,
  format: "epub",
  quality: "EPUB",
  hash: "abc123",
  addedAt: "2025-01-01T00:00:00Z",
  bookTitle: "Test Book",
  authorName: "Author",
  coverUrl: "",
}

const mockProgress: ReadingProgress = {
  id: 1,
  bookFileId: 1,
  userId: 1,
  position: "epubcfi(/6/2[cover]!/6)",
  percentage: 25.5,
  updatedAt: "2025-01-01T00:00:00Z",
}

const mockBookmark: Bookmark = {
  id: 1,
  bookFileId: 1,
  userId: 1,
  position: "epubcfi(/6/4!/4/2)",
  title: "My Bookmark",
  note: "Important passage",
  createdAt: "2025-01-01T00:00:00Z",
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe("useReaderBookFile", () => {
  it("fetches book file by ID", async () => {
    vi.mocked(readerApi.getBookFile).mockResolvedValue(mockBookFile)

    const { result } = renderHook(() => useReaderBookFile(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockBookFile)
    expect(readerApi.getBookFile).toHaveBeenCalledWith(1)
  })

  it("does not fetch when id is 0", () => {
    const { result } = renderHook(() => useReaderBookFile(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
    expect(readerApi.getBookFile).not.toHaveBeenCalled()
  })
})

describe("useReadingProgress", () => {
  it("fetches reading progress", async () => {
    vi.mocked(readerApi.getProgress).mockResolvedValue(mockProgress)

    const { result } = renderHook(() => useReadingProgress(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockProgress)
  })

  it("does not fetch when bookFileId is 0", () => {
    const { result } = renderHook(() => useReadingProgress(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useSaveProgress", () => {
  it("saves progress and updates cache", async () => {
    const updated = { ...mockProgress, percentage: 50 }
    vi.mocked(readerApi.saveProgress).mockResolvedValue(updated)

    const { result } = renderHook(() => useSaveProgress(1), { wrapper: createWrapper() })

    result.current.mutate({ position: "epubcfi(/6/4)", percentage: 50 })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(readerApi.saveProgress).toHaveBeenCalledWith(1, {
      position: "epubcfi(/6/4)",
      percentage: 50,
    })
  })
})

describe("useBookmarks", () => {
  it("fetches bookmarks for a book file", async () => {
    vi.mocked(readerApi.listBookmarks).mockResolvedValue([mockBookmark])

    const { result } = renderHook(() => useBookmarks(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([mockBookmark])
  })

  it("does not fetch when bookFileId is 0", () => {
    const { result } = renderHook(() => useBookmarks(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useCreateBookmark", () => {
  it("creates a bookmark", async () => {
    vi.mocked(readerApi.createBookmark).mockResolvedValue(mockBookmark)

    const { result } = renderHook(() => useCreateBookmark(1), { wrapper: createWrapper() })

    result.current.mutate({ position: "epubcfi(/6/4!/4/2)", title: "My Bookmark" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(readerApi.createBookmark).toHaveBeenCalledWith(1, {
      position: "epubcfi(/6/4!/4/2)",
      title: "My Bookmark",
    })
  })
})

describe("useDeleteBookmark", () => {
  it("deletes a bookmark", async () => {
    vi.mocked(readerApi.deleteBookmark).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteBookmark(1), { wrapper: createWrapper() })

    result.current.mutate(5)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(readerApi.deleteBookmark).toHaveBeenCalledWith(1, 5)
  })
})
