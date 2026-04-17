---
description: "Use when creating, editing, or reviewing Zustand v5 stores. Covers store structure, middleware ordering, immer mutations, persistence, and selector patterns."
applyTo: "web/src/store/**/*.ts"
---

# Zustand v5 — Store Conventions

## When to use Zustand

Use Zustand for **shared client state** that multiple components need to read:
- Auth state (`user`, `apiKey`, `isAuthenticated`) — synced from an XState machine
- UI preferences that must survive re-mounts (reader settings, theme)
- Cross-component state not owned by any single component tree

Do NOT use Zustand for server state — use TanStack Query for that.
Do NOT use Zustand for local component state — use `useState`/`useReducer`.

## Store Structure

Split state and actions into separate interfaces:

```ts
interface MyState {
  value: string
  count: number
}

interface MyActions {
  setValue: (v: string) => void
  increment: () => void
  reset: () => void
}

export const useMyStore = create<MyState & MyActions>()(/* middleware */)
```

## Middleware

### Standard middleware stack (with persistence):
```ts
create<State & Actions>()(
  devtools(
    persist(
      immer((set) => ({ /* state + actions */ })),
      {
        name: "bookaneer-<store-name>",   // localStorage key
        partialize: (state) => ({ key: state.key }), // persist only what's needed
      },
    ),
    { name: "<StoreName>Store", enabled: import.meta.env.DEV },
  ),
)
```

### Without persistence (e.g. ephemeral UI state):
```ts
create<State & Actions>()(
  devtools(
    immer((set) => ({ /* state + actions */ })),
    { name: "<StoreName>Store", enabled: import.meta.env.DEV },
  ),
)
```

### Middleware order matters
- `devtools` wraps everything — outermost
- `persist` comes next
- `immer` is innermost — closest to the state

## Persistence

- Use `partialize` to persist only what's needed — never persist derived state or functions
- Do NOT use Zustand `persist` if the data already has its own persistence mechanism
  (e.g. `ReaderSettingsStore` persists via `saveSettings()` from `readerConfig` — using
  both would duplicate writes)
- Persisted keys must be unique across stores: use `"bookaneer-<store-name>"` format

## Immer mutations

With the `immer` middleware, actions receive a mutable draft — mutate directly:

```ts
updateSettings: (updates) =>
  set((state) => {
    Object.assign(state, updates)   // ✅ mutate draft
    saveSettings({ ...state, ...updates })
  }),
```

To replace the whole state (e.g. `reset`), return a new object:
```ts
reset: () =>
  set(() => {
    saveSettings(DEFAULT_SETTINGS)
    return { ...DEFAULT_SETTINGS }  // ✅ return replaces state entirely
  }),
```

## Syncing with XState machines

When a Zustand store mirrors state from an XState machine, sync via subscription
in the Provider component — not inside the machine itself:

```ts
export function AuthProvider({ children }: { children: ReactNode }) {
  const actorRef = useActorRef(authMachine)
  const { setAuth, clearAuth } = useAuthStore()

  useEffect(() => {
    const sub = actorRef.subscribe((snapshot) => {
      if (snapshot.matches("authenticated") && snapshot.context.user) {
        setAuth(snapshot.context.user, snapshot.context.apiKey!)
      } else if (snapshot.matches("unauthenticated")) {
        clearAuth()
      }
    })
    return () => sub.unsubscribe()
  }, [actorRef, setAuth, clearAuth])

  // ...
}
```

XState is the source of truth for flow logic; Zustand is the read cache for components.

## Selectors

Always select the minimum slice — avoids unnecessary re-renders:

```ts
// ✅ Granular — only re-renders when isAuthenticated changes
const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

// ✅ Object selector — stable with shallow comparison (Zustand v5 default)
const settings = useReaderSettingsStore((s) => ({
  theme: s.theme,
  fontSize: s.fontSize,
}))

// ❌ Entire store — re-renders on any state change
const store = useAuthStore()
```

## File layout

```
web/src/store/<domain>/
└── <domain>.store.ts
```

One store file per domain. Export only the hook (`useMyStore`), not the store object itself.
