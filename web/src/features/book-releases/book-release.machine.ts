import { assign, fromPromise, setup } from "xstate"
import type { GrabResult } from "../../lib/schemas/wanted.schema"

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type GrabMeta =
  | { sourceType: "indexer"; guid?: string; seeders?: number; indexerId?: number; indexerName?: string }
  | { sourceType: "library" }

export type GrabFn = (
  bookId: number,
  downloadUrl: string,
  releaseTitle: string,
  size: number,
  meta?: GrabMeta,
) => Promise<GrabResult>

export type EnsureBookInLibraryFn = () => Promise<number>

interface PendingGrab {
  downloadUrl: string
  releaseTitle: string
  size: number
  meta?: GrabMeta
}

export interface BookReleaseMachineContext {
  formatFilter: string
  providerFilter: string
  languageFilter: string
  sortBy: string
  searchInResults: string
  isExpanded: boolean
  grabResult: GrabResult | null
  addError: string | null
  grabError: string | null
  ensureBookInLibraryFn: EnsureBookInLibraryFn | null
  grabFn: GrabFn | null
  pendingGrab: PendingGrab | null
}

export type BookReleaseMachineEvent =
  | { type: "START_SEARCH"; ensureBookInLibraryFn: EnsureBookInLibraryFn; grabFn: GrabFn }
  | { type: "CLOSE_SEARCH" }
  | { type: "EXPAND_SEARCH" }
  | { type: "SET_FORMAT_FILTER"; value: string }
  | { type: "SET_PROVIDER_FILTER"; value: string }
  | { type: "SET_LANGUAGE_FILTER"; value: string }
  | { type: "SET_SORT_BY"; value: string }
  | { type: "SET_SEARCH_IN_RESULTS"; value: string }
  | { type: "RESET_FILTERS" }
  | { type: "GRAB"; downloadUrl: string; releaseTitle: string; size: number; meta?: GrabMeta }

// ---------------------------------------------------------------------------
// Machine
// ---------------------------------------------------------------------------

export const bookReleaseMachine = setup({
  types: {
    context: {} as BookReleaseMachineContext,
    events: {} as BookReleaseMachineEvent,
  },
  actors: {
    grabActor: fromPromise<GrabResult, BookReleaseMachineContext>(async ({ input }) => {
      const { ensureBookInLibraryFn, grabFn, pendingGrab } = input
      if (!ensureBookInLibraryFn || !grabFn || !pendingGrab) {
        throw new Error("Grab not properly configured")
      }
      const bookId = await ensureBookInLibraryFn()
      return grabFn(bookId, pendingGrab.downloadUrl, pendingGrab.releaseTitle, pendingGrab.size, pendingGrab.meta)
    }),
  },
  actions: {
    initSearch: assign(({ event }) => {
      const e = event as { type: "START_SEARCH"; ensureBookInLibraryFn: EnsureBookInLibraryFn; grabFn: GrabFn }
      return {
        ensureBookInLibraryFn: e.ensureBookInLibraryFn,
        grabFn: e.grabFn,
        grabResult: null,
        addError: null,
        grabError: null,
        formatFilter: "all",
        providerFilter: "all",
        languageFilter: "all",
        sortBy: "score",
        searchInResults: "",
        isExpanded: false,
        pendingGrab: null,
      }
    }),
    resetSearch: assign({
      grabResult: null,
      addError: null,
      grabError: null,
      formatFilter: "all",
      providerFilter: "all",
      languageFilter: "all",
      sortBy: "score",
      searchInResults: "",
      isExpanded: false,
      pendingGrab: null,
      ensureBookInLibraryFn: null,
      grabFn: null,
    }),
    setPendingGrab: assign(({ event }) => {
      const e = event as { type: "GRAB"; downloadUrl: string; releaseTitle: string; size: number; meta?: GrabMeta }
      return {
        pendingGrab: { downloadUrl: e.downloadUrl, releaseTitle: e.releaseTitle, size: e.size, meta: e.meta },
        grabError: null,
      }
    }),
    setGrabResult: assign(({ event }) => ({
      grabResult: (event as unknown as { output: GrabResult }).output,
      grabError: null,
      pendingGrab: null,
    })),
    setGrabError: assign(({ event }) => {
      const e = event as unknown as { error: unknown }
      return {
        grabError: e.error instanceof Error ? e.error.message : "Failed to grab release",
        pendingGrab: null,
      }
    }),
    resetFilters: assign({
      formatFilter: "all",
      providerFilter: "all",
      languageFilter: "all",
      searchInResults: "",
    }),
    setFormatFilter: assign({ formatFilter: ({ event }) => (event as { value: string }).value }),
    setProviderFilter: assign({ providerFilter: ({ event }) => (event as { value: string }).value }),
    setLanguageFilter: assign({ languageFilter: ({ event }) => (event as { value: string }).value }),
    setSortBy: assign({ sortBy: ({ event }) => (event as { value: string }).value }),
    setSearchInResults: assign({ searchInResults: ({ event }) => (event as { value: string }).value }),
    expandSearch: assign({ isExpanded: true }),
  },
}).createMachine({
  id: "bookRelease",
  initial: "idle",
  context: {
    formatFilter: "all",
    providerFilter: "all",
    languageFilter: "all",
    sortBy: "score",
    searchInResults: "",
    isExpanded: false,
    grabResult: null,
    addError: null,
    grabError: null,
    ensureBookInLibraryFn: null,
    grabFn: null,
    pendingGrab: null,
  },
  states: {
    idle: {
      on: {
        START_SEARCH: { target: "searching", actions: "initSearch" },
      },
    },
    searching: {
      on: {
        CLOSE_SEARCH: { target: "idle", actions: "resetSearch" },
        EXPAND_SEARCH: { actions: "expandSearch" },
        SET_FORMAT_FILTER: { actions: "setFormatFilter" },
        SET_PROVIDER_FILTER: { actions: "setProviderFilter" },
        SET_LANGUAGE_FILTER: { actions: "setLanguageFilter" },
        SET_SORT_BY: { actions: "setSortBy" },
        SET_SEARCH_IN_RESULTS: { actions: "setSearchInResults" },
        RESET_FILTERS: { actions: "resetFilters" },
        GRAB: { target: "grabbing", actions: "setPendingGrab" },
      },
    },
    grabbing: {
      invoke: {
        id: "grab",
        src: "grabActor",
        input: ({ context }) => context,
        onDone: { target: "grabbed", actions: "setGrabResult" },
        onError: { target: "searching", actions: "setGrabError" },
      },
    },
    grabbed: {
      on: {
        GRAB: { target: "grabbing", actions: "setPendingGrab" },
        CLOSE_SEARCH: { target: "idle", actions: "resetSearch" },
        START_SEARCH: { target: "searching", actions: "initSearch" },
      },
    },
  },
})
