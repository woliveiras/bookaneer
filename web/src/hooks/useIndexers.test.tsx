import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { Indexer, IndexerOptions, SearchResponse, TestIndexerResponse } from "../lib/api"
import {
  useCreateIndexer,
  useDeleteIndexer,
  useIndexer,
  useIndexerOptions,
  useIndexers,
  useSearch,
  useTestIndexer,
  useUpdateIndexer,
  useUpdateIndexerOptions,
} from "./useIndexers"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    indexerApi: {
      list: vi.fn(),
      get: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      test: vi.fn(),
      getOptions: vi.fn(),
      updateOptions: vi.fn(),
      search: vi.fn(),
    },
  }
})

import { indexerApi } from "../lib/api"

const mockIndexer: Indexer = {
  id: 1,
  name: "Test Indexer",
  type: "newznab",
  baseUrl: "https://example.com",
  apiPath: "/api",
  apiKey: "test-key",
  categories: "7000,7020",
  priority: 25,
  enabled: true,
  enableRss: true,
  enableAutomaticSearch: true,
  enableInteractiveSearch: true,
  additionalParameters: "",
  minimumSeeders: 1,
  seedRatio: null,
  seedTime: null,
  createdAt: "2025-01-01T00:00:00Z",
  updatedAt: "2025-01-01T00:00:00Z",
}

const mockOptions: IndexerOptions = {
  minimumAge: 0,
  retention: 0,
  maximumSize: 0,
  rssSyncInterval: 15,
  preferIndexerFlags: false,
  availabilityDelay: 0,
  updatedAt: "2025-01-01T00:00:00Z",
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

describe("useIndexers", () => {
  it("fetches indexer list", async () => {
    vi.mocked(indexerApi.list).mockResolvedValue([mockIndexer])

    const { result } = renderHook(() => useIndexers(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([mockIndexer])
  })
})

describe("useIndexer", () => {
  it("fetches a single indexer by ID", async () => {
    vi.mocked(indexerApi.get).mockResolvedValue(mockIndexer)

    const { result } = renderHook(() => useIndexer(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockIndexer)
    expect(indexerApi.get).toHaveBeenCalledWith(1)
  })

  it("does not fetch when id is 0", () => {
    const { result } = renderHook(() => useIndexer(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
    expect(indexerApi.get).not.toHaveBeenCalled()
  })
})

describe("useCreateIndexer", () => {
  it("calls create", async () => {
    vi.mocked(indexerApi.create).mockResolvedValue(mockIndexer)

    const { result } = renderHook(() => useCreateIndexer(), { wrapper: createWrapper() })

    result.current.mutate({ name: "New", type: "newznab", baseUrl: "https://x.com", apiKey: "k" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(indexerApi.create).toHaveBeenCalled()
  })
})

describe("useUpdateIndexer", () => {
  it("calls update with id and data", async () => {
    const updated = { ...mockIndexer, name: "Updated" }
    vi.mocked(indexerApi.update).mockResolvedValue(updated)

    const { result } = renderHook(() => useUpdateIndexer(), { wrapper: createWrapper() })

    result.current.mutate({
      id: 1,
      data: { name: "Updated", type: "newznab", baseUrl: "https://x.com", apiKey: "k" },
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(indexerApi.update).toHaveBeenCalledWith(1, {
      name: "Updated",
      type: "newznab",
      baseUrl: "https://x.com",
      apiKey: "k",
    })
  })
})

describe("useDeleteIndexer", () => {
  it("calls delete with id", async () => {
    vi.mocked(indexerApi.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteIndexer(), { wrapper: createWrapper() })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(indexerApi.delete).toHaveBeenCalledWith(1)
  })
})

describe("useTestIndexer", () => {
  it("calls test with data", async () => {
    const response: TestIndexerResponse = { success: true, message: "OK" }
    vi.mocked(indexerApi.test).mockResolvedValue(response)

    const { result } = renderHook(() => useTestIndexer(), { wrapper: createWrapper() })

    result.current.mutate({ name: "Test", type: "newznab", baseUrl: "https://x.com", apiKey: "k" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
  })
})

describe("useSearch", () => {
  it("fetches search results when query provided", async () => {
    const response: SearchResponse = { results: [], total: 0 }
    vi.mocked(indexerApi.search).mockResolvedValue(response)

    const { result } = renderHook(() => useSearch({ q: "test book" }), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
  })

  it("does not fetch when no query fields provided", () => {
    const { result } = renderHook(() => useSearch({}), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
    expect(indexerApi.search).not.toHaveBeenCalled()
  })

  it("does not fetch when disabled", () => {
    const { result } = renderHook(() => useSearch({ q: "test" }, false), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useIndexerOptions", () => {
  it("fetches indexer options", async () => {
    vi.mocked(indexerApi.getOptions).mockResolvedValue(mockOptions)

    const { result } = renderHook(() => useIndexerOptions(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockOptions)
  })
})

describe("useUpdateIndexerOptions", () => {
  it("calls updateOptions", async () => {
    vi.mocked(indexerApi.updateOptions).mockResolvedValue(mockOptions)

    const { result } = renderHook(() => useUpdateIndexerOptions(), { wrapper: createWrapper() })

    const { updatedAt: _, ...input } = mockOptions
    result.current.mutate(input)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(indexerApi.updateOptions).toHaveBeenCalledWith(input)
  })
})
