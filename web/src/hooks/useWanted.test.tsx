import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type {
  ActiveCommand,
  BlocklistItem,
  BookSearchResponse,
  HistoryItem,
  QueueItem,
  WantedResponse,
} from "../lib/api"
import {
  useActiveCommands,
  useAddToBlocklist,
  useBlocklist,
  useDownloadQueue,
  useHistory,
  useManualGrab,
  useRecentCommands,
  useRemoveFromBlocklist,
  useRemoveFromQueue,
  useSearchBook,
  useWantedMissing,
} from "./useWanted"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    wantedApi: {
      getMissing: vi.fn(),
      searchBook: vi.fn(),
      manualGrab: vi.fn(),
      getActiveCommands: vi.fn(),
      getRecentCommands: vi.fn(),
    },
    queueApi: {
      list: vi.fn(),
      remove: vi.fn(),
    },
    historyApi: {
      list: vi.fn(),
    },
    blocklistApi: {
      list: vi.fn(),
      add: vi.fn(),
      remove: vi.fn(),
    },
  }
})

import { blocklistApi, historyApi, queueApi, wantedApi } from "../lib/api"

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

describe("useWantedMissing", () => {
  it("fetches missing books", async () => {
    const missing: WantedResponse = {
      page: 1,
      pageSize: 20,
      totalRecords: 0,
      sortKey: "title",
      sortDirection: "asc",
      records: [],
    }
    vi.mocked(wantedApi.getMissing).mockResolvedValue(missing)

    const { result } = renderHook(() => useWantedMissing(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(missing)
  })
})

describe("useSearchBook", () => {
  it("returns search results for a specific book", async () => {
    const response: BookSearchResponse = {
      results: [
        { title: "The Hobbit", source: "library", provider: "mock", format: "epub", size: 1024000, downloadUrl: "http://a.com/hobbit.epub" },
      ],
      noResults: false,
    }
    vi.mocked(wantedApi.searchBook).mockResolvedValue(response)

    const { result } = renderHook(() => useSearchBook(), { wrapper: createWrapper() })

    result.current.mutate(42)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wantedApi.searchBook).toHaveBeenCalledWith(42, expect.anything())
    expect(result.current.data?.noResults).toBe(false)
    expect(result.current.data?.results).toHaveLength(1)
  })

  it("returns noResults true when nothing is found", async () => {
    const response: BookSearchResponse = { results: [], noResults: true }
    vi.mocked(wantedApi.searchBook).mockResolvedValue(response)

    const { result } = renderHook(() => useSearchBook(), { wrapper: createWrapper() })

    result.current.mutate(99)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.noResults).toBe(true)
  })
})

describe("useDownloadQueue", () => {
  it("fetches queue items", async () => {
    const items: QueueItem[] = [
      {
        id: 1,
        bookId: 1,
        externalId: "ext-1",
        title: "Book",
        size: 1024,
        format: "epub",
        status: "downloading",
        progress: 50,
        downloadUrl: "",
        addedAt: "2025-01-01T00:00:00Z",
        bookTitle: "Book",
        clientName: "SABnzbd",
      },
    ]
    vi.mocked(queueApi.list).mockResolvedValue(items)

    const { result } = renderHook(() => useDownloadQueue(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(items)
  })
})

describe("useActiveCommands", () => {
  it("fetches active commands", async () => {
    const commands: ActiveCommand[] = [
      {
        id: "cmd-1",
        name: "BookSearch",
        status: "running",
        priority: 1,
        trigger: "manual",
        queuedAt: "2025-01-01T00:00:00Z",
      },
    ]
    vi.mocked(wantedApi.getActiveCommands).mockResolvedValue(commands)

    const { result } = renderHook(() => useActiveCommands(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(commands)
  })
})

describe("useManualGrab", () => {
  it("triggers a manual grab", async () => {
    const response: SearchCommandResponse = { commandId: "cmd-3", message: "Grab started" }
    vi.mocked(wantedApi.manualGrab).mockResolvedValue(response)

    const { result } = renderHook(() => useManualGrab(), { wrapper: createWrapper() })

    result.current.mutate({
      bookId: 1,
      downloadUrl: "http://example.com/book.epub",
      releaseTitle: "Book.epub",
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wantedApi.manualGrab).toHaveBeenCalledWith(
      { bookId: 1, downloadUrl: "http://example.com/book.epub", releaseTitle: "Book.epub" },
      expect.anything(),
    )
  })
})

describe("useRecentCommands", () => {
  it("fetches recent commands with default limit", async () => {
    const commands: ActiveCommand[] = [
      {
        id: "cmd-1",
        name: "BookSearch",
        status: "completed",
        priority: 1,
        trigger: "manual",
        queuedAt: "2025-01-01T00:00:00Z",
      },
    ]
    vi.mocked(wantedApi.getRecentCommands).mockResolvedValue(commands)

    const { result } = renderHook(() => useRecentCommands(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(commands)
    expect(wantedApi.getRecentCommands).toHaveBeenCalledWith(10)
  })

  it("fetches recent commands with custom limit", async () => {
    vi.mocked(wantedApi.getRecentCommands).mockResolvedValue([])

    const { result } = renderHook(() => useRecentCommands(5), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wantedApi.getRecentCommands).toHaveBeenCalledWith(5)
  })
})

describe("useRemoveFromQueue", () => {
  it("removes item and invalidates queue cache", async () => {
    vi.mocked(queueApi.remove).mockResolvedValue(undefined)

    const { result } = renderHook(() => useRemoveFromQueue(), { wrapper: createWrapper() })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(queueApi.remove).toHaveBeenCalledWith(1, expect.anything())
  })
})

describe("useHistory", () => {
  it("fetches history with params", async () => {
    const items: HistoryItem[] = [
      {
        id: 1,
        eventType: "grabbed",
        sourceTitle: "Book.epub",
        quality: "EPUB",
        data: {},
        date: "2025-01-01T00:00:00Z",
      },
    ]
    vi.mocked(historyApi.list).mockResolvedValue(items)

    const params = { limit: 10 }
    const { result } = renderHook(() => useHistory(params), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(historyApi.list).toHaveBeenCalledWith(params)
  })
})

describe("useBlocklist", () => {
  it("fetches blocklist", async () => {
    const items: BlocklistItem[] = [
      {
        id: 1,
        bookId: 1,
        sourceTitle: "Blocked",
        quality: "EPUB",
        reason: "Bad quality",
        date: "2025-01-01T00:00:00Z",
        bookTitle: "Blocked",
        authorName: "Author",
      },
    ]
    vi.mocked(blocklistApi.list).mockResolvedValue(items)

    const { result } = renderHook(() => useBlocklist(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(items)
  })
})

describe("useAddToBlocklist", () => {
  it("adds to blocklist", async () => {
    vi.mocked(blocklistApi.add).mockResolvedValue(undefined)

    const { result } = renderHook(() => useAddToBlocklist(), { wrapper: createWrapper() })

    result.current.mutate({ bookId: 1, sourceTitle: "Bad Book" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(blocklistApi.add).toHaveBeenCalledWith(
      { bookId: 1, sourceTitle: "Bad Book" },
      expect.anything(),
    )
  })
})

describe("useRemoveFromBlocklist", () => {
  it("removes from blocklist", async () => {
    vi.mocked(blocklistApi.remove).mockResolvedValue(undefined)

    const { result } = renderHook(() => useRemoveFromBlocklist(), { wrapper: createWrapper() })

    result.current.mutate(5)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(blocklistApi.remove).toHaveBeenCalledWith(5, expect.anything())
  })
})
