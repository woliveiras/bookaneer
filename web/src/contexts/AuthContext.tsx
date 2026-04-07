import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from "react"
import {
  type User,
  authApi,
  getStoredApiKey,
  setStoredApiKey,
  clearStoredApiKey,
} from "../lib/api"

interface AuthContextValue {
  user: User | null
  apiKey: string | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (apiKey: string) => Promise<void>
  loginWithCredentials: (username: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null)
  const [apiKey, setApiKey] = useState<string | null>(getStoredApiKey)
  const [isLoading, setIsLoading] = useState(true)

  // Check stored API key on mount
  useEffect(() => {
    const storedKey = getStoredApiKey()
    if (storedKey) {
      authApi
        .login(storedKey)
        .then((userData) => {
          setUser(userData)
          setApiKey(storedKey)
        })
        .catch(() => {
          // Invalid key, clear it
          clearStoredApiKey()
          setApiKey(null)
        })
        .finally(() => setIsLoading(false))
    } else {
      setIsLoading(false)
    }
  }, [])

  const login = useCallback(async (newApiKey: string) => {
    const userData = await authApi.login(newApiKey)
    setStoredApiKey(newApiKey)
    setApiKey(newApiKey)
    setUser(userData)
  }, [])

  const loginWithCredentials = useCallback(async (username: string, password: string) => {
    const response = await authApi.loginWithCredentials(username, password)
    setStoredApiKey(response.apiKey)
    setApiKey(response.apiKey)
    setUser(response.user)
  }, [])

  const logout = useCallback(() => {
    clearStoredApiKey()
    setApiKey(null)
    setUser(null)
  }, [])

  const value: AuthContextValue = {
    user,
    apiKey,
    isAuthenticated: !!apiKey,
    isLoading,
    login,
    loginWithCredentials,
    logout,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
