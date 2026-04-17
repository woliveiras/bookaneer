import { useActorRef } from "@xstate/react"
import { createContext, type ReactNode, useContext, useEffect } from "react"
import type { ActorRefFrom } from "xstate"
import { authMachine } from "./auth.machine"
import { useAuthStore } from "../../store/auth/auth.store"

export type AuthActorRef = ActorRefFrom<typeof authMachine>
export const AuthActorContext = createContext<AuthActorRef | null>(null)

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const actorRef = useActorRef(authMachine)
  const { setAuth, clearAuth } = useAuthStore()

  // Sync auth machine state into Zustand store
  useEffect(() => {
    const sub = actorRef.subscribe((snapshot) => {
      if (
        snapshot.matches("authenticated") &&
        snapshot.context.user &&
        snapshot.context.apiKey
      ) {
        setAuth(snapshot.context.user, snapshot.context.apiKey)
      } else if (snapshot.matches("unauthenticated")) {
        clearAuth()
      }
    })
    return () => sub.unsubscribe()
  }, [actorRef, setAuth, clearAuth])

  return <AuthActorContext.Provider value={actorRef}>{children}</AuthActorContext.Provider>
}

/** Access the auth machine actor to send events (LOGIN, LOGOUT, etc.) */
export function useAuthActor(): AuthActorRef {
  const ctx = useContext(AuthActorContext)
  if (!ctx) throw new Error("useAuthActor must be used within AuthProvider")
  return ctx
}
