import { create } from "zustand"
import { devtools, persist } from "zustand/middleware"
import type { User } from "../../lib/schemas/auth.schema"

interface AuthState {
  user: User | null
  apiKey: string | null
  isAuthenticated: boolean
}

interface AuthActions {
  setAuth: (user: User, apiKey: string) => void
  clearAuth: () => void
}

export const useAuthStore = create<AuthState & AuthActions>()(
  devtools(
    persist(
      (set) => ({
        user: null,
        apiKey: null,
        isAuthenticated: false,
        setAuth: (user, apiKey) => set({ user, apiKey, isAuthenticated: true }),
        clearAuth: () => set({ user: null, apiKey: null, isAuthenticated: false }),
      }),
      {
        name: "bookaneer-auth",
        // Only persist apiKey — user object will be re-validated on restore
        partialize: (state) => ({ apiKey: state.apiKey }),
      },
    ),
    { name: "AuthStore", enabled: import.meta.env.DEV },
  ),
)
