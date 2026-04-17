import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type {
  DigitalLibrarySearchResponse,
  MetadataAuthorResult,
  MetadataBook,
  MetadataBookResult,
  MetadataSearchResponse,
} from "../lib/api"
import {
  useDigitalLibrarySearch,
  useMetadataBook,
  useMetadataSearchAuthors,
  useMetadataSearchBooks,
} from "./useMetadata"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    metadataApi: {
      searchAuthors: vi.fn(),
      searchBooks: vi.fn(),
      getBook: vi.fn(),
    },
    digitalLibraryApi: {
      search: vi.fn(),
    },
  }
})

import { digitalLibraryApi, metadataApi } from "../lib/api"

const mockBookResult: MetadataBookResult = {
  foreignId: "OL1W",
  title: "Test Book",
  authors: ["Test Author"],
  publishedYear: 2024,
  provider: "openlibrary",
}

const mockMetadataBook: MetadataBook = {
  foreignId: "OL1W",
  title: "Test Book",
  authors: ["Test Author"],
  description: "A great book",
  provider: "openlibrary",
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

describe("useMetadataSearchAuthors", () => {
  it("searches authors when query >= 2 chars", async () => {
    const response: MetadataSearchResponse<MetadataAuthorResult> = {
      results: [{ foreignId: "OL1A", name: "Test Author", worksCount: 5, provider: "openlibrary" }],
      total: 1,
    }
    vi.mocked(metadataApi.searchAuthors).mockResolvedValue(response)

    const { result } = renderHook(() => useMetadataSearchAuthors("test"), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
  })

  it("does not search when query < 2 chars", () => {
    const { result } = renderHook(() => useMetadataSearchAuthors("a"), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })

  it("does not search when disabled", () => {
    const { result } = renderHook(() => useMetadataSearchAuthors("test", false), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useMetadataSearchBooks", () => {
  it("searches books when query >= 2 chars", async () => {
    const response: MetadataSearchResponse<MetadataBookResult> = {
      results: [mockBookResult],
      total: 1,
    }
    vi.mocked(metadataApi.searchBooks).mockResolvedValue(response)

    const { result } = renderHook(() => useMetadataSearchBooks("test"), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
  })

  it("does not search when query < 2 chars", () => {
    const { result } = renderHook(() => useMetadataSearchBooks("a"), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useMetadataBook", () => {
  it("fetches book details", async () => {
    vi.mocked(metadataApi.getBook).mockResolvedValue(mockMetadataBook)

    const { result } = renderHook(() => useMetadataBook("OL1W"), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockMetadataBook)
  })

  it("does not fetch when foreignId is empty", () => {
    const { result } = renderHook(() => useMetadataBook(""), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useDigitalLibrarySearch", () => {
  it("searches digital libraries when query >= 2 chars", async () => {
    const response: DigitalLibrarySearchResponse = { results: [], total: 0 }
    vi.mocked(digitalLibraryApi.search).mockResolvedValue(response)

    const { result } = renderHook(() => useDigitalLibrarySearch("test book"), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
  })

  it("does not search when query < 2 chars", () => {
    const { result } = renderHook(() => useDigitalLibrarySearch("a"), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })

  it("does not search when disabled", () => {
    const { result } = renderHook(() => useDigitalLibrarySearch("test", false), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

