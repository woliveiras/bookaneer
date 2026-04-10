import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { CreateGrabInput, DownloadClient, Grab, QueueItem } from "../lib/api"
import {
  useClientQueue,
  useCreateDownloadClient,
  useCreateGrab,
  useDeleteDownloadClient,
  useDownloadClient,
  useDownloadClients,
  useGrabs,
  useQueue,
  useSendGrab,
  useTestDownloadClient,
  useUpdateDownloadClient,
} from "./useDownload"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    downloadClientApi: {
      list: vi.fn(),
      get: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      test: vi.fn(),
    },
    queueApi: {
      list: vi.fn(),
      listByClient: vi.fn(),
    },
    grabApi: {
      list: vi.fn(),
      create: vi.fn(),
      send: vi.fn(),
    },
  }
})

import { downloadClientApi, grabApi, queueApi } from "../lib/api"

const mockClient: DownloadClient = {
  id: 1,
  name: "SABnzbd",
  type: "sabnzbd",
  host: "localhost",
  port: 8080,
  useTls: false,
  username: "",
  password: "",
  apiKey: "test-key",
  category: "books",
  recentPriority: 0,
  olderPriority: 0,
  removeCompletedAfter: 0,
  enabled: true,
  priority: 1,
  nzbFolder: "",
  torrentFolder: "",
  watchFolder: "",
  createdAt: "2025-01-01T00:00:00Z",
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

describe("useDownloadClients", () => {
  it("fetches client list", async () => {
    vi.mocked(downloadClientApi.list).mockResolvedValue([mockClient])

    const { result } = renderHook(() => useDownloadClients(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([mockClient])
  })
})

describe("useDownloadClient", () => {
  it("fetches single client by ID", async () => {
    vi.mocked(downloadClientApi.get).mockResolvedValue(mockClient)

    const { result } = renderHook(() => useDownloadClient(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockClient)
  })

  it("does not fetch when id is 0", () => {
    const { result } = renderHook(() => useDownloadClient(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
  })
})

describe("useCreateDownloadClient", () => {
  it("creates client and invalidates cache", async () => {
    vi.mocked(downloadClientApi.create).mockResolvedValue(mockClient)

    const { result } = renderHook(() => useCreateDownloadClient(), { wrapper: createWrapper() })

    result.current.mutate({ name: "SABnzbd", type: "sabnzbd" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(downloadClientApi.create).toHaveBeenCalledWith(
      { name: "SABnzbd", type: "sabnzbd" },
      expect.anything(),
    )
  })
})

describe("useUpdateDownloadClient", () => {
  it("updates client", async () => {
    vi.mocked(downloadClientApi.update).mockResolvedValue(mockClient)

    const { result } = renderHook(() => useUpdateDownloadClient(), { wrapper: createWrapper() })

    result.current.mutate({ id: 1, data: { name: "Updated", type: "sabnzbd" } })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(downloadClientApi.update).toHaveBeenCalledWith(1, { name: "Updated", type: "sabnzbd" })
  })
})

describe("useDeleteDownloadClient", () => {
  it("deletes client", async () => {
    vi.mocked(downloadClientApi.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteDownloadClient(), { wrapper: createWrapper() })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(downloadClientApi.delete).toHaveBeenCalledWith(1, expect.anything())
  })
})

describe("useTestDownloadClient", () => {
  it("tests client connection", async () => {
    vi.mocked(downloadClientApi.test).mockResolvedValue({ success: true, message: "OK" })

    const { result } = renderHook(() => useTestDownloadClient(), { wrapper: createWrapper() })

    result.current.mutate({ name: "Test", type: "sabnzbd" as const })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(downloadClientApi.test).toHaveBeenCalledWith(
      { name: "Test", type: "sabnzbd" },
      expect.anything(),
    )
  })
})

describe("useQueue", () => {
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

    const { result } = renderHook(() => useQueue(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(items)
  })
})

describe("useGrabs", () => {
  it("fetches grabs with filter params", async () => {
    const grabs: Grab[] = [
      {
        id: 1,
        bookId: 1,
        indexerId: 1,
        releaseTitle: "Book.epub",
        downloadUrl: "http://example.com/book",
        size: 1024,
        quality: "EPUB",
        clientId: 1,
        downloadId: "dl-1",
        status: "sent",
        errorMessage: "",
        grabbedAt: "2025-01-01T00:00:00Z",
      },
    ]
    vi.mocked(grabApi.list).mockResolvedValue(grabs)

    const params = { bookId: 1 }
    const { result } = renderHook(() => useGrabs(params), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(grabApi.list).toHaveBeenCalledWith(params)
  })
})

describe("useCreateGrab", () => {
  it("creates grab and invalidates caches", async () => {
    const mockGrab: Grab = {
      id: 1,
      bookId: 1,
      indexerId: 1,
      releaseTitle: "Book.epub",
      downloadUrl: "http://example.com/book",
      size: 1024,
      quality: "EPUB",
      clientId: 1,
      downloadId: "dl-1",
      status: "pending",
      errorMessage: "",
      grabbedAt: "2025-01-01T00:00:00Z",
    }
    vi.mocked(grabApi.create).mockResolvedValue(mockGrab)

    const { result } = renderHook(() => useCreateGrab(), { wrapper: createWrapper() })

    const input: CreateGrabInput = {
      bookId: 1,
      indexerId: 1,
      releaseTitle: "Book.epub",
      downloadUrl: "http://example.com/book",
      clientId: 1,
    }
    result.current.mutate(input)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(grabApi.create).toHaveBeenCalledWith(input, expect.anything())
  })
})

describe("useClientQueue", () => {
  it("fetches queue items for a specific client", async () => {
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
    vi.mocked(queueApi.listByClient).mockResolvedValue(items)

    const { result } = renderHook(() => useClientQueue(1), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(items)
    expect(queueApi.listByClient).toHaveBeenCalledWith(1)
  })

  it("does not fetch when clientId is 0", () => {
    const { result } = renderHook(() => useClientQueue(0), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe("idle")
    expect(queueApi.listByClient).not.toHaveBeenCalled()
  })
})

describe("useSendGrab", () => {
  it("sends grab and invalidates caches", async () => {
    const sentGrab: Grab = {
      id: 1,
      bookId: 1,
      indexerId: 1,
      releaseTitle: "Book.epub",
      downloadUrl: "http://example.com/book",
      size: 1024,
      quality: "EPUB",
      clientId: 1,
      downloadId: "dl-1",
      status: "sent",
      errorMessage: "",
      grabbedAt: "2025-01-01T00:00:00Z",
    }
    vi.mocked(grabApi.send).mockResolvedValue(sentGrab)

    const { result } = renderHook(() => useSendGrab(), { wrapper: createWrapper() })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(grabApi.send).toHaveBeenCalledWith(1, expect.anything())
  })
})
