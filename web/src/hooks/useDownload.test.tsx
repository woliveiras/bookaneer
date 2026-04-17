import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { DownloadClient } from "../lib/api"
import {
  useCreateDownloadClient,
  useDeleteDownloadClient,
  useDownloadClients,
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
  }
})

import { downloadClientApi } from "../lib/api"

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

