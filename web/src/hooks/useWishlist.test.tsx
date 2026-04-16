import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { Book, PaginatedResponse } from "../lib/api"
import { useAddToWishlist, useRemoveFromWishlist, useWishlist } from "./useWishlist"

vi.mock("../lib/api/wishlist", () => ({
  wishlistApi: {
    list: vi.fn(),
    add: vi.fn(),
    remove: vi.fn(),
  },
}))

import { wishlistApi } from "../lib/api/wishlist"

const mockBook: Book = {
  id: 1,
  authorId: 5,
  title: "The Name of the Wind",
  sortTitle: "Name of the Wind",
  foreignId: "OL12345W",
  isbn: "",
  isbn13: "9780756404741",
  releaseDate: "2007-03-27",
  overview: "",
  imageUrl: "https://covers.example.com/1.jpg",
  pageCount: 662,
  inWishlist: true,
  addedAt: "2025-06-01T00:00:00Z",
  updatedAt: "2025-06-01T00:00:00Z",
  authorName: "Patrick Rothfuss",
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

describe("useWishlist", () => {
  it("fetches wishlist books", async () => {
    const response: PaginatedResponse<Book> = {
      records: [mockBook],
      totalRecords: 1,
    }
    vi.mocked(wishlistApi.list).mockResolvedValue(response)

    const { result } = renderHook(() => useWishlist(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
    expect(wishlistApi.list).toHaveBeenCalledTimes(1)
  })

  it("returns empty list when wishlist is empty", async () => {
    const response: PaginatedResponse<Book> = { records: [], totalRecords: 0 }
    vi.mocked(wishlistApi.list).mockResolvedValue(response)

    const { result } = renderHook(() => useWishlist(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.totalRecords).toBe(0)
    expect(result.current.data?.records).toHaveLength(0)
  })

  it("exposes error on failure", async () => {
    vi.mocked(wishlistApi.list).mockRejectedValue(new Error("Network error"))

    const { result } = renderHook(() => useWishlist(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.error).toBeInstanceOf(Error)
  })
})

describe("useAddToWishlist", () => {
  it("adds a book to the wishlist", async () => {
    vi.mocked(wishlistApi.add).mockResolvedValue(mockBook)

    const { result } = renderHook(() => useAddToWishlist(), { wrapper: createWrapper() })

    result.current.mutate({
      title: "The Name of the Wind",
      authors: ["Patrick Rothfuss"],
      foreignId: "OL12345W",
      isbn13: "9780756404741",
      imageUrl: "https://covers.example.com/1.jpg",
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wishlistApi.add).toHaveBeenCalledWith(
      {
        title: "The Name of the Wind",
        authors: ["Patrick Rothfuss"],
        foreignId: "OL12345W",
        isbn13: "9780756404741",
        imageUrl: "https://covers.example.com/1.jpg",
      },
      expect.anything(),
    )
    expect(result.current.data).toEqual(mockBook)
  })

  it("returns error when add fails", async () => {
    vi.mocked(wishlistApi.add).mockRejectedValue(new Error("Conflict"))

    const { result } = renderHook(() => useAddToWishlist(), { wrapper: createWrapper() })

    result.current.mutate({ title: "Fail Book", authors: [] })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.error).toBeInstanceOf(Error)
  })
})

describe("useRemoveFromWishlist", () => {
  it("removes a book from the wishlist by id", async () => {
    vi.mocked(wishlistApi.remove).mockResolvedValue(undefined)

    const { result } = renderHook(() => useRemoveFromWishlist(), { wrapper: createWrapper() })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(wishlistApi.remove).toHaveBeenCalledWith(1, expect.anything())
  })

  it("returns error when remove fails", async () => {
    vi.mocked(wishlistApi.remove).mockRejectedValue(new Error("Not Found"))

    const { result } = renderHook(() => useRemoveFromWishlist(), { wrapper: createWrapper() })

    result.current.mutate(999)

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.error).toBeInstanceOf(Error)
  })
})
