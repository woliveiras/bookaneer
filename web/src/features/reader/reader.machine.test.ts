import { createActor } from "xstate"
import { describe, expect, it } from "vitest"
import { readerMachine } from "./reader.machine"
import type { ReaderBookFile } from "../../lib/schemas/reader.schema"

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const mockBookFile: ReaderBookFile = {
  id: 1,
  bookId: 10,
  path: "/books/clean-code.epub",
  relativePath: "clean-code.epub",
  format: "epub",
  size: 2048000,
  quality: "good",
  hash: "abc123",
  addedAt: "2024-01-01T00:00:00Z",
}

const mockToc = [
  { label: "Introduction", href: "intro.html", id: "intro", subitems: [] },
  { label: "Chapter 1", href: "ch1.html", id: "ch1", subitems: [] },
]

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("readerMachine", () => {
  it("starts in idle state with default context", () => {
    const actor = createActor(readerMachine)
    actor.start()
    const snapshot = actor.getSnapshot()
    expect(snapshot.value).toBe("idle")
    expect(snapshot.context.isLoading).toBe(false)
    expect(snapshot.context.error).toBeNull()
    expect(snapshot.context.toc).toEqual([])
    expect(snapshot.context.progress).toBe(0)
    actor.stop()
  })

  it("transitions to loading on INIT and sets isLoading", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    const snapshot = actor.getSnapshot()
    expect(snapshot.value).toBe("loading")
    expect(snapshot.context.isLoading).toBe(true)
    expect(snapshot.context.error).toBeNull()
    actor.stop()
  })

  it("transitions to ready on LOAD_SUCCESS and clears isLoading", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({ type: "LOAD_SUCCESS" })
    const snapshot = actor.getSnapshot()
    expect(snapshot.value).toBe("ready")
    expect(snapshot.context.isLoading).toBe(false)
    actor.stop()
  })

  it("transitions to failed on LOAD_ERROR and stores the error message", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({ type: "LOAD_ERROR", error: "Failed to fetch book content" })
    const snapshot = actor.getSnapshot()
    expect(snapshot.value).toBe("failed")
    expect(snapshot.context.error).toBe("Failed to fetch book content")
    expect(snapshot.context.isLoading).toBe(false)
    actor.stop()
  })

  it("updates location and progress in loading state via LOCATION_UPDATED", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({
      type: "LOCATION_UPDATED",
      location: "Chapter 1",
      cfi: "epubcfi(/6/4[chap01]!/4/2/1:0)",
      progress: 0.15,
    })
    const ctx = actor.getSnapshot().context
    expect(ctx.currentLocation).toBe("Chapter 1")
    expect(ctx.currentCfi).toBe("epubcfi(/6/4[chap01]!/4/2/1:0)")
    expect(ctx.progress).toBe(0.15)
    actor.stop()
  })

  it("updates location and progress in ready state via LOCATION_UPDATED", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({ type: "LOAD_SUCCESS" })
    actor.send({
      type: "LOCATION_UPDATED",
      location: "Chapter 5",
      cfi: "epubcfi(/6/14[chap05]!/4/2/1:0)",
      progress: 0.42,
    })
    const ctx = actor.getSnapshot().context
    expect(ctx.currentLocation).toBe("Chapter 5")
    expect(ctx.progress).toBe(0.42)
    actor.stop()
  })

  it("updates toc in loading state via TOC_UPDATED", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({ type: "TOC_UPDATED", toc: mockToc })
    expect(actor.getSnapshot().context.toc).toEqual(mockToc)
    actor.stop()
  })

  it("updates toc in ready state via TOC_UPDATED", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({ type: "LOAD_SUCCESS" })
    actor.send({ type: "TOC_UPDATED", toc: mockToc })
    expect(actor.getSnapshot().context.toc).toEqual(mockToc)
    actor.stop()
  })

  it("re-initializes from ready state on INIT (e.g. theme change)", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({ type: "LOAD_SUCCESS" })
    expect(actor.getSnapshot().value).toBe("ready")

    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    expect(actor.getSnapshot().value).toBe("loading")
    expect(actor.getSnapshot().context.isLoading).toBe(true)
    actor.stop()
  })

  it("re-initializes from failed state on INIT", () => {
    const actor = createActor(readerMachine)
    actor.start()
    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    actor.send({ type: "LOAD_ERROR", error: "Network error" })
    expect(actor.getSnapshot().value).toBe("failed")

    actor.send({ type: "INIT", bookFile: mockBookFile, bookFileId: 1 })
    expect(actor.getSnapshot().value).toBe("loading")
    actor.stop()
  })
})
