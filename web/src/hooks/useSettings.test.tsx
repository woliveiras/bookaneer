import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { renderHook, waitFor } from "@testing-library/react"
import type { ReactNode } from "react"
import { beforeEach, describe, expect, it, vi } from "vitest"
import type { GeneralSettings } from "../lib/api"
import { useGeneralSettings } from "./useSettings"

vi.mock("../lib/api", async () => {
  const actual = await vi.importActual<typeof import("../lib/api")>("../lib/api")
  return {
    ...actual,
    settingsApi: {
      getGeneral: vi.fn(),
    },
  }
})

import { settingsApi } from "../lib/api"

const mockSettings: GeneralSettings = {
  apiKey: "test-api-key",
  bindAddress: "0.0.0.0",
  port: 8787,
  dataDir: "/data",
  libraryDir: "/library",
  logLevel: "info",
  customProvidersEnabled: true,
  customProvidersActive: [
    {
      name: "provider-test",
      domain: "example.org",
      formatHint: "pdf",
    },
  ],
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

describe("useGeneralSettings", () => {
  it("fetches general settings", async () => {
    vi.mocked(settingsApi.getGeneral).mockResolvedValue(mockSettings)

    const { result } = renderHook(() => useGeneralSettings(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(mockSettings)
    expect(settingsApi.getGeneral).toHaveBeenCalled()
  })
})
