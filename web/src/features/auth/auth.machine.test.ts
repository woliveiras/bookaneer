import { createActor } from "xstate"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { authMachine } from "./auth.machine"

// Mock the API and storage modules
vi.mock("../../lib/api", () => ({
  authApi: {
    login: vi.fn(),
    loginWithCredentials: vi.fn(),
  },
  getStoredApiKey: vi.fn(),
  setStoredApiKey: vi.fn(),
  clearStoredApiKey: vi.fn(),
}))

import { authApi, clearStoredApiKey, getStoredApiKey, setStoredApiKey } from "../../lib/api"

const mockUser = {
  id: 1,
  username: "admin",
  role: "admin",
  createdAt: "2024-01-01T00:00:00Z",
}

beforeEach(() => {
  vi.clearAllMocks()
  vi.mocked(getStoredApiKey).mockReturnValue(null)
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe("authMachine", () => {
  describe("checkingSession", () => {
    it("starts in unauthenticated when no stored key", () => {
      vi.mocked(getStoredApiKey).mockReturnValue(null)
      const actor = createActor(authMachine)
      actor.start()
      expect(actor.getSnapshot().matches("unauthenticated")).toBe(true)
      actor.stop()
    })

    it("transitions to authenticating when stored key exists", () => {
      vi.mocked(getStoredApiKey).mockReturnValue("stored-key")
      vi.mocked(authApi.login).mockResolvedValue(mockUser)
      const actor = createActor(authMachine)
      actor.start()
      // Immediately after start, should be in authenticating (async actor not resolved yet)
      expect(actor.getSnapshot().matches("authenticating")).toBe(true)
      actor.stop()
    })
  })

  describe("unauthenticated", () => {
    it("transitions to loggingIn on LOGIN_WITH_CREDENTIALS event", () => {
      vi.mocked(getStoredApiKey).mockReturnValue(null)
      vi.mocked(authApi.loginWithCredentials).mockResolvedValue({
        user: mockUser,
        apiKey: "new-api-key",
      })
      const actor = createActor(authMachine)
      actor.start()

      actor.send({ type: "LOGIN_WITH_CREDENTIALS", username: "admin", password: "secret" })

      expect(actor.getSnapshot().matches("loggingIn")).toBe(true)
      actor.stop()
    })
  })

  describe("loggingIn", () => {
    it("transitions to authenticated on successful login", async () => {
      vi.mocked(getStoredApiKey).mockReturnValue(null)
      const mockResponse = { user: mockUser, apiKey: "new-api-key" }
      vi.mocked(authApi.loginWithCredentials).mockResolvedValue(mockResponse)

      const actor = createActor(authMachine)
      actor.start()
      actor.send({ type: "LOGIN_WITH_CREDENTIALS", username: "admin", password: "secret" })

      // Wait for loginWithCredentials to be called
      await vi.waitFor(() => expect(authApi.loginWithCredentials).toHaveBeenCalled())
      // Wait for the promise to resolve and state to update
      await vi.waitFor(() => actor.getSnapshot().context.user !== null)

      expect(actor.getSnapshot().context.user?.username).toBe("admin")
      expect(actor.getSnapshot().context.apiKey).toBe("new-api-key")
      expect(setStoredApiKey).toHaveBeenCalledWith("new-api-key")
      actor.stop()
    })

    it("transitions to unauthenticated with error on failed login", async () => {
      vi.mocked(getStoredApiKey).mockReturnValue(null)
      vi.mocked(authApi.loginWithCredentials).mockRejectedValue(new Error("Invalid credentials"))

      const actor = createActor(authMachine)
      actor.start()
      actor.send({ type: "LOGIN_WITH_CREDENTIALS", username: "admin", password: "wrong" })

      // Wait for mock to be called, then for error to appear in context
      await vi.waitFor(() => expect(authApi.loginWithCredentials).toHaveBeenCalled())
      await vi.waitFor(() => actor.getSnapshot().context.error !== null)

      expect(actor.getSnapshot().context.error).toBe("Invalid credentials")
      actor.stop()
    })
  })

  describe("authenticated", () => {
    it("transitions to unauthenticated on LOGOUT", async () => {
      vi.mocked(getStoredApiKey).mockReturnValue(null)
      vi.mocked(authApi.loginWithCredentials).mockResolvedValue({
        user: mockUser,
        apiKey: "key",
      })

      const actor = createActor(authMachine)
      actor.start()
      actor.send({ type: "LOGIN_WITH_CREDENTIALS", username: "admin", password: "secret" })
      await vi.waitFor(() => expect(authApi.loginWithCredentials).toHaveBeenCalled())
      await vi.waitFor(() => actor.getSnapshot().context.user !== null)

      actor.send({ type: "LOGOUT" })

      await vi.waitFor(() => actor.getSnapshot().context.user === null)
      expect(clearStoredApiKey).toHaveBeenCalled()
      actor.stop()
    })
  })
})
