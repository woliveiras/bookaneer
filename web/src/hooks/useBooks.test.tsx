import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { Book, BookWithEditions, PaginatedResponse } from "../lib/api"
import { useBook, useBooks, useCreateBook, useDeleteBook, useUpdateBook } from "./useBooks"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    bookApi: {
      list: vi.fn(),
      get: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    },
  }
})

import { bookApi } from "../lib/api"

const mockBook: Book = {
  id: 1,
  authorId: 10,
  title: "Test Book",
  sortTitle: "Test Book",
  foreignId: "fid-1",
  isbn: "1234567890",
  isbn13: "1234567890123",
  releaseDate: "2024-01-01",
  overview: "A test book",
  imageUrl: "",
  pageCount: 200,
  monitored: true,
  addedAt: "2025-01-01T00:00:00Z",
  updatedAt: "2025-01-01T00:00:00Z",
}

const mockBookWithEditions: BookWithEditions = {
  ...mockBook,
  editions: [
    {
      id: 1,
      bookId: 1,
      foreignId: "ed-1",
      title: "Test Edition",
      isbn: "1234567890",
      isbn13: "1234567890123",
      format: "epub",
      publisher: "Publisher",
      releaseDate: "2024-01-01",
      pageCount: 200,
      language: "en",
      monitored: true,
    },
  ],
  files: [],
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

describe("useBooks", () => {
  it("fetches book list", async () => {
    const response: PaginatedResponse<Book> = {
      records: [mockBook],
      totalRecords: 1,
    }
    vi.mocked(bookApi.list).mockResolvedValue(response)

    const { result } = renderHook(() => useBooks(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
    expect(bookApi.list).toHaveBeenCalledWith(undefined)
  })

  it("passes filter params to API", async () => {
    const response: PaginatedResponse<Book> = { records: [], totalRecords: 0 }
    vi.mocked(bookApi.list).mockResolvedValue(response)

    const params = { authorId: 10, monitored: true }
    const { result } = renderHook(() => useBooks(params), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(bookApi.list).toHaveBeenCalledWith(params)
  })
})

describe("useBook", () => {
  it("fetches a single book by ID", async () => {
    vi.mocked(bookApi.get).mockResolvedValue(mockBookWithEditions)

    const { result } = renderHook(() => useBook(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockBookWithEditions)
    expect(bookApi.get).toHaveBeenCalledWith(1)
  })

  it("does not fetch when id is 0", () => {
    const { result } = renderHook(() => useBook(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
    expect(bookApi.get).not.toHaveBeenCalled()
  })
})

describe("useCreateBook", () => {
  it("calls create and invalidates cache", async () => {
    vi.mocked(bookApi.create).mockResolvedValue(mockBook)

    const { result } = renderHook(() => useCreateBook(), { wrapper: createWrapper() })

    result.current.mutate({ authorId: 10, title: "Test Book" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(bookApi.create).toHaveBeenCalledWith({ authorId: 10, title: "Test Book" })
  })
})

describe("useUpdateBook", () => {
  it("calls update with id and data", async () => {
    const updated = { ...mockBook, title: "Updated" }
    vi.mocked(bookApi.update).mockResolvedValue(updated)

    const { result } = renderHook(() => useUpdateBook(), { wrapper: createWrapper() })

    result.current.mutate({ id: 1, data: { title: "Updated" } })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(bookApi.update).toHaveBeenCalledWith(1, { title: "Updated" })
  })
})

describe("useDeleteBook", () => {
  it("calls delete with id", async () => {
    vi.mocked(bookApi.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteBook(), { wrapper: createWrapper() })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(bookApi.delete).toHaveBeenCalledWith(1)
  })
})
