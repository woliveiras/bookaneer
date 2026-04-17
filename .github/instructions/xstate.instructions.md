---
description: "Use when creating, editing, or reviewing XState v5 state machines. Covers createMachine vs setup(), onDone/onError event typing, actor patterns, and testing with createActor."
applyTo: "web/src/features/**/*.machine.ts, web/src/features/**/*.machine.test.ts"
---

# XState v5 — State Machine Conventions

## When to use XState

Use XState for **flows with multiple lifecycle states** and side effects:
- Auth: `checkingSession → authenticating → authenticated / unauthenticated → loggingIn`
- Reader lifecycle: `idle → loading → ready / failed`
- Multi-step processes with loading/error/success states
- Workflows with guarded transitions or parallel states

Do NOT use XState for simple toggle state or preferences — use Zustand for that.

## Machine Definition

### Prefer `createMachine(config, implementations)` when using `onDone`/`onError`

`setup().createMachine()` has a known issue: actions defined in `setup.actions` receive
the event typed as your union type. `DoneActorEvent` is NOT in that union, so XState
cannot pass `event.output` to the action correctly.

**❌ Avoid for machines with `onDone`/`onError`:**
```ts
export const myMachine = setup({
  actions: {
    setResult: assign({ result: ({ event }) => event.output }), // event.output is undefined at runtime
  },
}).createMachine({ ... })
```

**✅ Use `createMachine` directly with inline `assign` and `unknown` cast:**
```ts
export const myMachine = createMachine(
  {
    // config
    states: {
      loading: {
        invoke: {
          src: "fetchData",
          onDone: {
            target: "ready",
            actions: assign(({ event }: { event: unknown }) => {
              const e = event as { output: { data: Data } }
              return { data: e.output.data, error: null }
            }),
          },
          onError: {
            target: "failed",
            actions: assign(({ event }: { event: unknown }) => {
              const e = event as { error: unknown }
              return { error: e.error instanceof Error ? e.error.message : "Unknown error" }
            }),
          },
        },
      },
    },
  },
  { actors: { fetchData: fetchDataActor } },
)
```

**Exception:** `setup().createMachine()` is fine when the machine has no `invoke`/`onDone`/`onError` —
pure event-driven machines with synchronous actions work correctly.

## Typing

Always declare `types` at the top of the machine config:

```ts
export const myMachine = createMachine({
  types: {
    context: {} as MyContext,
    events: {} as MyEvent,
  },
  // ...
})
```

This enables TypeScript inference for `send()` call sites and `useSelector` callbacks.

## Actors

- Use `fromPromise` for async operations (API calls, async init)
- Use `fromCallback` for event-based / DOM listeners (no return value needed)
- For actors that never resolve (placeholder), return a no-op cleanup from `fromCallback`:

```ts
actors: {
  myPlaceholderActor: fromCallback<MyEvent, MyInput>((_params) => {
    return () => {} // cleanup
  }),
}
```

- Pass runtime dependencies via `input` — never close over mutable state in the actor factory:

```ts
invoke: {
  src: "grabActor",
  input: ({ context }) => context, // pass what the actor needs
}
```

## Providing actors for tests

Override actors in tests with `machine.provide()`:

```ts
const actor = createActor(
  myMachine.provide({
    actors: {
      fetchData: fromPromise(async () => mockResult),
    },
  }),
)
```

## Consuming machines in components

```ts
import { useActorRef, useSelector } from "@xstate/react"

// In a provider component — creates the actor once
const actorRef = useActorRef(myMachine)

// In consumers — subscribe to only what they need (avoids unnecessary re-renders)
const isLoading = useSelector(actorRef, (s) => s.matches("loading"))
const error = useSelector(actorRef, (s) => s.context.error)

// Dispatch events
actorRef.send({ type: "RETRY" })
```

Expose `actorRef` via React context so consumers don't re-create the machine:

```ts
export const MyActorContext = createContext<MyActorRef | null>(null)

export function MyProvider({ children }: { children: ReactNode }) {
  const actorRef = useActorRef(myMachine)
  return <MyActorContext.Provider value={actorRef}>{children}</MyActorContext.Provider>
}
```

## Testing

Always use `createActor` (not `interpret` — removed in v5):

```ts
const actor = createActor(myMachine)
actor.start()
actor.send({ type: "SUBMIT" })
expect(actor.getSnapshot().value).toBe("submitting")
actor.stop()
```

### Async state assertions

**❌ Unreliable — `vi.waitFor(() => snapshot.matches("X"))` can miss transient states:**
```ts
await vi.waitFor(() => actor.getSnapshot().matches("authenticated"))
```

**✅ Reliable — wait for observable side effects, then check context:**
```ts
// Wait for the actor/mock to have been called
await vi.waitFor(() => expect(authApi.login).toHaveBeenCalled())
// Then check the context value (not the state name)
await vi.waitFor(() => actor.getSnapshot().context.user !== null)
```

### Always stop actors after tests

```ts
afterEach(() => actor.stop())
// or inside the test:
actor.stop()
```

## File layout

```
web/src/features/<domain>/
├── <domain>.machine.ts       # machine definition + exported types
└── <domain>.machine.test.ts  # tests with createActor
```

Export types used by consumers (e.g. `GrabMeta`, `GrabFn`) from the machine file,
not from hook files — the machine is the source of truth for the domain's types.
