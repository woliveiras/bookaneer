import { renderHook, waitFor } from "@testing-library/react"
import { describe, it, expect, vi, beforeEach } from "vitest"
import type { ReactNode } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import {
  useWantedMissing,
  useSearchAllMissing,
  useSearchBook,
  useDownloadQueue,
  useActiveCommands,
  useRemoveFromQueue,
  useHistory,
  useBlocklist,
  useAddToBlocklist,
  useRemoveFromBlocklist,
} from "./useWanted"
import type { WantedResponse, SearchCommandResponse, QueueItem, ActiveCommand, HistoryItem, BlocklistItem } from "../lib/api"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    wantedApi: {
      getMissing: vi.fn(),
      searchAllMissing: vi.fn(),
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

import { wantedApi, queueApi, historyApi, blocklistApi } from "../lib/api"

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

describe("useSearchAllMissing", () => {
  it("triggers search and resolves", async () => {
    const response: SearchCommandResponse = { commandId: "cmd-1", message: "Search started" }
    vi.mocked(wantedApi.searchAllMissing).mockResolvedValue(response)

    const { result } = renderHook(() => useSearchAllMissing(), { wrapper: createWrapper() })

    result.current.mutate()

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wantedApi.searchAllMissing).toHaveBeenCalled()
  })
})

describe("useSearchBook", () => {
  it("triggers search for specific book", async () => {
    const response: SearchCommandResponse = { commandId: "cmd-2", message: "Search started" }
    vi.mocked(wantedApi.searchBook).mockResolvedValue(response)

    const { result } = renderHook(() => useSearchBook(), { wrapper: createWrapper() })

    result.current.mutate(42)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wantedApi.searchBook).toHaveBeenCalledWith(42, expect.anything())
  })
})

describe("useDownloadQueue", () => {
  it("fetches queue items", async () => {
    const items: QueueItem[] = [{
      id: 1, bookId: 1, externalId: "ext-1", title: "Book", size: 1024,
      format: "epub", status: "downloading", progress: 50, downloadUrl: "",
      addedAt: "2025-01-01T00:00:00Z", bookTitle: "Book", clientName: "SABnzbd",
    }]
    vi.mocked(queueApi.list).mockResolvedValue(items)

    const { result } = renderHook(() => useDownloadQueue(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(items)
  })
})

describe("useActiveCommands", () => {
  it("fetches active commands", async () => {
    const commands: ActiveCommand[] = [{
      id: "cmd-1", name: "BookSearch", status: "running",
      priority: 1, trigger: "manual", queuedAt: "2025-01-01T00:00:00Z",
    }]
    vi.mocked(wantedApi.getActiveCommands).mockResolvedValue(commands)

    const { result } = renderHook(() => useActiveCommands(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(commands)
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
    const items: HistoryItem[] = [{
      id: 1, eventType: "grabbed", sourceTitle: "Book.epub",
      quality: "EPUB", data: {}, date: "2025-01-01T00:00:00Z",
    }]
    vi.mocked(historyApi.list).mockResolvedValue(items)

    const params = { limit: 10 }
    const { result } = renderHook(() => useHistory(params), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(historyApi.list).toHaveBeenCalledWith(params)
  })
})

describe("useBlocklist", () => {
  it("fetches blocklist", async () => {
    const items: BlocklistItem[] = [{
      id: 1, bookId: 1, sourceTitle: "Blocked", quality: "EPUB",
      reason: "Bad quality", date: "2025-01-01T00:00:00Z",
      bookTitle: "Blocked", authorName: "Author",
    }]
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
    expect(blocklistApi.add).toHaveBeenCalledWith({ bookId: 1, sourceTitle: "Bad Book" }, expect.anything())
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
