import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type {
  DigitalLibrarySearchResponse,
  MetadataAuthor,
  MetadataAuthorResult,
  MetadataBook,
  MetadataBookResult,
  MetadataSearchResponse,
} from "../lib/api"
import {
  useDigitalLibraryProviders,
  useDigitalLibrarySearch,
  useMetadataAuthor,
  useMetadataBook,
  useMetadataISBN,
  useMetadataProviders,
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
      getAuthor: vi.fn(),
      getBook: vi.fn(),
      lookupISBN: vi.fn(),
      getProviders: vi.fn(),
    },
    digitalLibraryApi: {
      search: vi.fn(),
      getProviders: vi.fn(),
    },
  }
})

import { digitalLibraryApi, metadataApi } from "../lib/api"

const mockAuthorResult: MetadataAuthorResult = {
  foreignId: "OL1A",
  name: "Test Author",
  worksCount: 5,
  provider: "openlibrary",
}

const mockBookResult: MetadataBookResult = {
  foreignId: "OL1W",
  title: "Test Book",
  authors: ["Test Author"],
  publishedYear: 2024,
  provider: "openlibrary",
}

const mockMetadataAuthor: MetadataAuthor = {
  foreignId: "OL1A",
  name: "Test Author",
  sortName: "Author, Test",
  bio: "A great author",
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
      results: [mockAuthorResult],
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

describe("useMetadataAuthor", () => {
  it("fetches author details", async () => {
    vi.mocked(metadataApi.getAuthor).mockResolvedValue(mockMetadataAuthor)

    const { result } = renderHook(() => useMetadataAuthor("OL1A"), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockMetadataAuthor)
    expect(metadataApi.getAuthor).toHaveBeenCalledWith("OL1A", undefined)
  })

  it("passes provider param", async () => {
    vi.mocked(metadataApi.getAuthor).mockResolvedValue(mockMetadataAuthor)

    const { result } = renderHook(() => useMetadataAuthor("OL1A", "openlibrary"), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(metadataApi.getAuthor).toHaveBeenCalledWith("OL1A", "openlibrary")
  })

  it("does not fetch when foreignId is empty", () => {
    const { result } = renderHook(() => useMetadataAuthor(""), {
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

describe("useMetadataISBN", () => {
  it("looks up book by ISBN", async () => {
    vi.mocked(metadataApi.lookupISBN).mockResolvedValue(mockMetadataBook)

    const { result } = renderHook(() => useMetadataISBN("1234567890"), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockMetadataBook)
  })

  it("does not fetch when ISBN < 10 chars", () => {
    const { result } = renderHook(() => useMetadataISBN("12345"), {
      wrapper: createWrapper(),
    })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useMetadataProviders", () => {
  it("fetches metadata providers", async () => {
    const response = { providers: ["openlibrary", "google"] }
    vi.mocked(metadataApi.getProviders).mockResolvedValue(response)

    const { result } = renderHook(() => useMetadataProviders(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
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

describe("useDigitalLibraryProviders", () => {
  it("fetches digital library providers", async () => {
    const response = { providers: ["annas_archive", "libgen"] }
    vi.mocked(digitalLibraryApi.getProviders).mockResolvedValue(response)

    const { result } = renderHook(() => useDigitalLibraryProviders(), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
  })
})
