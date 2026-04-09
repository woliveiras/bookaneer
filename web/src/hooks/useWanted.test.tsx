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
    const missing = { books: [], total: 0 }
    vi.mocked(wantedApi.getMissing).mockResolvedValue(missing)

    const { result } = renderHook(() => useWantedMissing(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(missing)
  })
})

describe("useSearchAllMissing", () => {
  it("triggers search and resolves", async () => {
    vi.mocked(wantedApi.searchAllMissing).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSearchAllMissing(), { wrapper: createWrapper() })

    result.current.mutate()

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wantedApi.searchAllMissing).toHaveBeenCalled()
  })
})

describe("useSearchBook", () => {
  it("triggers search for specific book", async () => {
    vi.mocked(wantedApi.searchBook).mockResolvedValue(undefined)

    const { result } = renderHook(() => useSearchBook(), { wrapper: createWrapper() })

    result.current.mutate(42)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wantedApi.searchBook).toHaveBeenCalledWith(42, expect.anything())
  })
})

describe("useDownloadQueue", () => {
  it("fetches queue items", async () => {
    const items = [{ id: 1, title: "Book", status: "downloading" }]
    vi.mocked(queueApi.list).mockResolvedValue(items)

    const { result } = renderHook(() => useDownloadQueue(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(items)
  })
})

describe("useActiveCommands", () => {
  it("fetches active commands", async () => {
    const commands = [{ id: "cmd-1", status: "running" }]
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
    const items = [{ id: 1, eventType: "grabbed" }]
    vi.mocked(historyApi.list).mockResolvedValue(items)

    const params = { limit: 10 }
    const { result } = renderHook(() => useHistory(params), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(historyApi.list).toHaveBeenCalledWith(params)
  })
})

describe("useBlocklist", () => {
  it("fetches blocklist", async () => {
    const items = [{ id: 1, title: "Blocked" }]
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

    result.current.mutate({ title: "Bad Book", indexerId: 1, guid: "abc" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(blocklistApi.add).toHaveBeenCalledWith({ title: "Bad Book", indexerId: 1, guid: "abc" }, expect.anything())
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
