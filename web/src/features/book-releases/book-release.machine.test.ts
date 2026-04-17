import { createActor, fromPromise } from "xstate"
import { describe, expect, it, vi } from "vitest"
import { bookReleaseMachine } from "./book-release.machine"
import type { BookReleaseMachineContext, EnsureBookInLibraryFn, GrabFn, GrabMeta } from "./book-release.machine"
import type { GrabResult } from "../../lib/schemas/wanted.schema"

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const mockGrabResult: GrabResult = {
  bookId: 42,
  title: "Clean Code",
  source: "test-indexer",
  providerName: "TestProvider",
  format: "epub",
  size: 1024000,
  downloadId: "dl-abc123",
  clientName: "SABnzbd",
}

const makeFns = (grabResult: GrabResult = mockGrabResult) => {
  const ensureBookInLibraryFn: EnsureBookInLibraryFn = vi.fn().mockResolvedValue(42)
  const grabFn: GrabFn = vi.fn().mockResolvedValue(grabResult)
  return { ensureBookInLibraryFn, grabFn }
}

const pendingGrabEvent = {
  type: "GRAB" as const,
  downloadUrl: "https://example.com/book.epub",
  releaseTitle: "Clean Code",
  size: 1024000,
  meta: { sourceType: "indexer" } as GrabMeta,
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("bookReleaseMachine", () => {
  it("starts in idle state with default context", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    const snapshot = actor.getSnapshot()
    expect(snapshot.value).toBe("idle")
    expect(snapshot.context.grabResult).toBeNull()
    expect(snapshot.context.grabError).toBeNull()
    expect(snapshot.context.formatFilter).toBe("all")
    actor.stop()
  })

  it("transitions to searching on START_SEARCH and stores fns in context", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    const { ensureBookInLibraryFn, grabFn } = makeFns()

    actor.send({ type: "START_SEARCH", ensureBookInLibraryFn, grabFn })

    const snapshot = actor.getSnapshot()
    expect(snapshot.value).toBe("searching")
    expect(snapshot.context.ensureBookInLibraryFn).toBe(ensureBookInLibraryFn)
    expect(snapshot.context.grabFn).toBe(grabFn)
    actor.stop()
  })

  it("resets filters when going idle → searching via START_SEARCH", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    const { ensureBookInLibraryFn, grabFn } = makeFns()

    // First search session: start, set filters, close
    actor.send({ type: "START_SEARCH", ensureBookInLibraryFn, grabFn })
    actor.send({ type: "SET_FORMAT_FILTER", value: "epub" })
    actor.send({ type: "SET_LANGUAGE_FILTER", value: "pt" })
    actor.send({ type: "CLOSE_SEARCH" })
    expect(actor.getSnapshot().value).toBe("idle")

    // Second search session: filters should reset to defaults
    actor.send({ type: "START_SEARCH", ensureBookInLibraryFn, grabFn })
    const ctx = actor.getSnapshot().context
    expect(ctx.formatFilter).toBe("all")
    expect(ctx.languageFilter).toBe("all")
    expect(ctx.sortBy).toBe("score")
    actor.stop()
  })

  it("updates format filter in searching state", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send({ type: "SET_FORMAT_FILTER", value: "epub" })
    expect(actor.getSnapshot().context.formatFilter).toBe("epub")
    actor.stop()
  })

  it("updates language filter in searching state", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send({ type: "SET_LANGUAGE_FILTER", value: "en" })
    expect(actor.getSnapshot().context.languageFilter).toBe("en")
    actor.stop()
  })

  it("updates sort order in searching state", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send({ type: "SET_SORT_BY", value: "size-asc" })
    expect(actor.getSnapshot().context.sortBy).toBe("size-asc")
    actor.stop()
  })

  it("resets all filters on RESET_FILTERS", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send({ type: "SET_FORMAT_FILTER", value: "pdf" })
    actor.send({ type: "SET_LANGUAGE_FILTER", value: "fr" })
    actor.send({ type: "SET_SORT_BY", value: "year-desc" })
    actor.send({ type: "RESET_FILTERS" })
    const ctx = actor.getSnapshot().context
    expect(ctx.formatFilter).toBe("all")
    expect(ctx.languageFilter).toBe("all")
    actor.stop()
  })

  it("marks isExpanded on EXPAND_SEARCH", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send({ type: "EXPAND_SEARCH" })
    expect(actor.getSnapshot().context.isExpanded).toBe(true)
    actor.stop()
  })

  it("transitions back to idle on CLOSE_SEARCH", () => {
    const actor = createActor(bookReleaseMachine)
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send({ type: "CLOSE_SEARCH" })
    expect(actor.getSnapshot().value).toBe("idle")
    expect(actor.getSnapshot().context.ensureBookInLibraryFn).toBeNull()
    actor.stop()
  })

  it("transitions to grabbing on GRAB and stores pendingGrab", () => {
    const actor = createActor(
      bookReleaseMachine.provide({
        actors: {
          // Actor that never resolves so we can inspect the grabbing state
          grabActor: fromPromise<GrabResult, BookReleaseMachineContext>(async () => {
            await new Promise<never>(() => {})
            return mockGrabResult
          }),
        },
      }),
    )
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send(pendingGrabEvent)

    const snapshot = actor.getSnapshot()
    expect(snapshot.value).toBe("grabbing")
    expect(snapshot.context.pendingGrab?.downloadUrl).toBe(pendingGrabEvent.downloadUrl)
    expect(snapshot.context.pendingGrab?.releaseTitle).toBe(pendingGrabEvent.releaseTitle)
    actor.stop()
  })

  it("transitions to grabbed with result on successful grab", async () => {
    const actor = createActor(
      bookReleaseMachine.provide({
        actors: {
          grabActor: fromPromise<GrabResult, BookReleaseMachineContext>(async () => mockGrabResult),
        },
      }),
    )
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send(pendingGrabEvent)

    await vi.waitFor(() => actor.getSnapshot().value === "grabbed")
    expect(actor.getSnapshot().context.grabResult).toEqual(mockGrabResult)
    expect(actor.getSnapshot().context.grabError).toBeNull()
    actor.stop()
  })

  it("returns to searching with error on failed grab", async () => {
    const actor = createActor(
      bookReleaseMachine.provide({
        actors: {
          grabActor: fromPromise<GrabResult, BookReleaseMachineContext>(async () => {
            throw new Error("Connection refused")
          }),
        },
      }),
    )
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send(pendingGrabEvent)

    await vi.waitFor(() => actor.getSnapshot().context.grabError !== null)
    expect(actor.getSnapshot().value).toBe("searching")
    expect(actor.getSnapshot().context.grabError).toBe("Connection refused")
    actor.stop()
  })

  it("allows grabbing again from grabbed state", async () => {
    const actor = createActor(
      bookReleaseMachine.provide({
        actors: {
          grabActor: fromPromise<GrabResult, BookReleaseMachineContext>(async () => mockGrabResult),
        },
      }),
    )
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send(pendingGrabEvent)
    await vi.waitFor(() => actor.getSnapshot().value === "grabbed")

    // Grab again — should go back to grabbing
    actor.send(pendingGrabEvent)
    expect(actor.getSnapshot().value).toBe("grabbing")
    actor.stop()
  })

  it("allows closing search from grabbed state", async () => {
    const actor = createActor(
      bookReleaseMachine.provide({
        actors: {
          grabActor: fromPromise<GrabResult, BookReleaseMachineContext>(async () => mockGrabResult),
        },
      }),
    )
    actor.start()
    actor.send({ type: "START_SEARCH", ...makeFns() })
    actor.send(pendingGrabEvent)
    await vi.waitFor(() => actor.getSnapshot().value === "grabbed")

    actor.send({ type: "CLOSE_SEARCH" })
    expect(actor.getSnapshot().value).toBe("idle")
    actor.stop()
  })
})
