import { renderHook, waitFor } from "@testing-library/react"
import { describe, it, expect, vi, beforeEach } from "vitest"
import type { ReactNode } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { useAuthors, useAuthor, useAuthorStats, useCreateAuthor, useUpdateAuthor, useDeleteAuthor } from "./useAuthors"
import type { Author, AuthorStats, PaginatedResponse } from "../lib/api"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    authorApi: {
      list: vi.fn(),
      get: vi.fn(),
      getStats: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    },
  }
})

import { authorApi } from "../lib/api"

const mockAuthor: Author = {
  id: 1,
  name: "Test Author",
  sortName: "Author, Test",
  foreignId: "fid-1",
  overview: "",
  imageUrl: "",
  status: "continuing",
  monitored: true,
  path: "/books/test-author",
  addedAt: "2025-01-01T00:00:00Z",
  updatedAt: "2025-01-01T00:00:00Z",
}

const mockStats: AuthorStats = {
  bookCount: 5,
  bookFileCount: 3,
  missingBooks: 2,
  totalSizeBytes: 1024000,
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

describe("useAuthors", () => {
  it("fetches author list", async () => {
    const response: PaginatedResponse<Author> = {
      items: [mockAuthor],
      total: 1,
      limit: 20,
      offset: 0,
    }
    vi.mocked(authorApi.list).mockResolvedValue(response)

    const { result } = renderHook(() => useAuthors(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
    expect(authorApi.list).toHaveBeenCalledWith(undefined)
  })

  it("passes filter params to API", async () => {
    const response: PaginatedResponse<Author> = { items: [], total: 0, limit: 20, offset: 0 }
    vi.mocked(authorApi.list).mockResolvedValue(response)

    const params = { monitored: true, search: "test" }
    const { result } = renderHook(() => useAuthors(params), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(authorApi.list).toHaveBeenCalledWith(params)
  })
})

describe("useAuthor", () => {
  it("fetches a single author by ID", async () => {
    vi.mocked(authorApi.get).mockResolvedValue(mockAuthor)

    const { result } = renderHook(() => useAuthor(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockAuthor)
    expect(authorApi.get).toHaveBeenCalledWith(1)
  })

  it("does not fetch when id is 0", () => {
    const { result } = renderHook(() => useAuthor(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
    expect(authorApi.get).not.toHaveBeenCalled()
  })
})

describe("useAuthorStats", () => {
  it("fetches stats for author", async () => {
    vi.mocked(authorApi.getStats).mockResolvedValue(mockStats)

    const { result } = renderHook(() => useAuthorStats(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockStats)
  })
})

describe("useCreateAuthor", () => {
  it("calls create and invalidates query cache", async () => {
    vi.mocked(authorApi.create).mockResolvedValue(mockAuthor)

    const wrapper = createWrapper()
    const { result } = renderHook(() => useCreateAuthor(), { wrapper })

    result.current.mutate({ name: "Test Author" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(authorApi.create).toHaveBeenCalledWith({ name: "Test Author" })
  })
})

describe("useUpdateAuthor", () => {
  it("calls update with id and data", async () => {
    const updated = { ...mockAuthor, name: "Updated" }
    vi.mocked(authorApi.update).mockResolvedValue(updated)

    const { result } = renderHook(() => useUpdateAuthor(), { wrapper: createWrapper() })

    result.current.mutate({ id: 1, data: { name: "Updated" } })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(authorApi.update).toHaveBeenCalledWith(1, { name: "Updated" })
  })
})

describe("useDeleteAuthor", () => {
  it("calls delete with id", async () => {
    vi.mocked(authorApi.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteAuthor(), { wrapper: createWrapper() })

    result.current.mutate({ id: 1 })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(authorApi.delete).toHaveBeenCalledWith(1, undefined)
  })
})
