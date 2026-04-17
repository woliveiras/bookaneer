import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { RootFolder } from "../lib/api"
import {
  useCreateRootFolder,
  useDeleteRootFolder,
  useRootFolders,
  useUpdateRootFolder,
} from "./useRootFolders"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    rootFolderApi: {
      list: vi.fn(),
      get: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    },
  }
})

import { rootFolderApi } from "../lib/api"

const mockFolder: RootFolder = {
  id: 1,
  path: "/media/books",
  name: "Books",
  defaultQualityProfileId: 1,
  freeSpace: 100000000,
  totalSpace: 500000000,
  authorCount: 10,
  accessible: true,
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

describe("useRootFolders", () => {
  it("fetches root folder list", async () => {
    vi.mocked(rootFolderApi.list).mockResolvedValue([mockFolder])

    const { result } = renderHook(() => useRootFolders(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual([mockFolder])
  })
})

describe("useCreateRootFolder", () => {
  it("creates a root folder", async () => {
    vi.mocked(rootFolderApi.create).mockResolvedValue(mockFolder)

    const { result } = renderHook(() => useCreateRootFolder(), { wrapper: createWrapper() })

    result.current.mutate({ path: "/media/books", name: "Books" })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(rootFolderApi.create).toHaveBeenCalledWith(
      { path: "/media/books", name: "Books" },
      expect.anything(),
    )
  })
})

describe("useUpdateRootFolder", () => {
  it("updates a root folder", async () => {
    const updated = { ...mockFolder, name: "My Books" }
    vi.mocked(rootFolderApi.update).mockResolvedValue(updated)

    const { result } = renderHook(() => useUpdateRootFolder(), { wrapper: createWrapper() })

    result.current.mutate({ id: 1, data: { name: "My Books" } })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(rootFolderApi.update).toHaveBeenCalledWith(1, { name: "My Books" })
  })
})

describe("useDeleteRootFolder", () => {
  it("deletes a root folder", async () => {
    vi.mocked(rootFolderApi.delete).mockResolvedValue(undefined)

    const { result } = renderHook(() => useDeleteRootFolder(), { wrapper: createWrapper() })

    result.current.mutate(1)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(rootFolderApi.delete).toHaveBeenCalledWith(1, expect.anything())
  })
})
