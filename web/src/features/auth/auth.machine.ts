import { assign, createMachine, fromPromise } from "xstate"
import {
  authApi,
  clearStoredApiKey,
  getStoredApiKey,
  setStoredApiKey,
} from "../../lib/api"
import type { User } from "../../lib/schemas/auth.schema"

interface AuthContext {
  user: User | null
  apiKey: string | null
  error: string | null
}

type AuthEvent =
  | { type: "LOGIN"; apiKey: string }
  | { type: "LOGIN_WITH_CREDENTIALS"; username: string; password: string }
  | { type: "LOGOUT" }

const loginWithApiKeyActor = fromPromise(
  async ({ input }: { input: { apiKey: string } }) => {
    const user = await authApi.login(input.apiKey)
    return { user, apiKey: input.apiKey }
  },
)

const loginWithCredentialsActor = fromPromise(
  async ({ input }: { input: { username: string; password: string } }) => {
    const response = await authApi.loginWithCredentials(input.username, input.password)
    return { user: response.user, apiKey: response.apiKey }
  },
)

export const authMachine = createMachine(
  {
    id: "auth",
    initial: "checkingSession",
    types: {
      context: {} as AuthContext,
      events: {} as AuthEvent,
    },
    context: {
      user: null,
      apiKey: null,
      error: null,
    },
    states: {
      checkingSession: {
        always: [
          {
            guard: () => !!getStoredApiKey(),
            target: "authenticating",
            actions: assign({ apiKey: () => getStoredApiKey() }),
          },
          { target: "unauthenticated" },
        ],
      },

      authenticating: {
        invoke: {
          src: "loginWithApiKey",
          input: ({ context }: { context: AuthContext }) => ({ apiKey: context.apiKey ?? "" }),
          onDone: {
            target: "authenticated",
            actions: assign(({ event }: { event: unknown }) => {
              const e = event as { output: { user: User | null; apiKey: string } }
              return {
                user: e.output.user,
                apiKey: e.output.apiKey,
                error: null as string | null,
              }
            }),
          },
          onError: {
            target: "unauthenticated",
            actions: [
              assign({ apiKey: null, user: null, error: null }),
              () => clearStoredApiKey(),
            ],
          },
        },
      },

      authenticated: {
        on: {
          LOGOUT: {
            target: "unauthenticated",
            actions: [
              assign({ user: null, apiKey: null, error: null }),
              () => clearStoredApiKey(),
            ],
          },
        },
      },

      unauthenticated: {
        on: {
          LOGIN: {
            target: "authenticating",
            actions: [
              assign(({ event }: { event: { type: "LOGIN"; apiKey: string } }) => ({
                apiKey: event.apiKey,
                error: null as string | null,
              })),
              ({ event }: { event: { type: "LOGIN"; apiKey: string } }) => setStoredApiKey(event.apiKey),
            ],
          },
          LOGIN_WITH_CREDENTIALS: {
            target: "loggingIn",
            actions: assign({ error: null }),
          },
        },
      },

      loggingIn: {
        invoke: {
          src: "loginWithCredentials",
          input: ({ event }: { event: AuthEvent }) => {
            if (event.type !== "LOGIN_WITH_CREDENTIALS") throw new Error("Unexpected event")
            return { username: event.username, password: event.password }
          },
          onDone: {
            target: "authenticated",
            actions: [
              assign(({ event }: { event: unknown }) => {
                const e = event as { output: { user: User | null; apiKey: string } }
                return {
                  user: e.output.user,
                  apiKey: e.output.apiKey,
                  error: null as string | null,
                }
              }),
              ({ event }: { event: unknown }) => {
                const e = event as { output: { apiKey: string } }
                setStoredApiKey(e.output.apiKey)
              },
            ],
          },
          onError: {
            target: "unauthenticated",
            actions: assign(({ event }: { event: unknown }) => {
              const e = event as { error: unknown }
              return {
                error: e.error instanceof Error ? e.error.message : "Authentication failed",
              }
            }),
          },
        },
      },
    },
  },
  {
    actors: {
      loginWithApiKey: loginWithApiKeyActor,
      loginWithCredentials: loginWithCredentialsActor,
    },
  },
)
