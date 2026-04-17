import { assign, fromCallback, setup } from "xstate"
import type { FoliateView, TocItem } from "../../components/reader/readerConfig"
import type { ReaderBookFile } from "../../lib/schemas/reader.schema"

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface ReaderMachineContext {
  isLoading: boolean
  error: string | null
  currentLocation: string
  currentCfi: string
  progress: number
  toc: TocItem[]
}

type ReaderMachineEvent =
  | { type: "INIT"; bookFile: ReaderBookFile; bookFileId: number; initialCfi?: string }
  | { type: "LOAD_SUCCESS" }
  | { type: "LOAD_ERROR"; error: string }
  | { type: "TOC_UPDATED"; toc: TocItem[] }
  | { type: "LOCATION_UPDATED"; location: string; cfi: string; progress: number }
  | { type: "NAVIGATE_PREV" }
  | { type: "NAVIGATE_NEXT" }
  | { type: "NAVIGATE_TO"; target: string }

interface ReaderMachineInput {
  containerRef: React.RefObject<HTMLDivElement | null>
  viewRef: React.MutableRefObject<FoliateView | null>
  getApplyStyles: () => () => void
  getSaveProgress: () => (cfi: string, progress: number) => void
  getInitialCfi: () => string | undefined
}

// ---------------------------------------------------------------------------
// Machine
// ---------------------------------------------------------------------------

export const readerMachine = setup({
  types: {
    context: {} as ReaderMachineContext,
    events: {} as ReaderMachineEvent,
  },
  actors: {
    /**
     * No-op loader actor — actual init logic stays in useReaderCore.
     * The machine tracks state transitions via events sent from the hook.
     */
    readerLoader: fromCallback<ReaderMachineEvent, ReaderMachineInput>(
      (_params) => {
        return () => {}
      },
    ),
  },
  actions: {
    setLoading: assign({ isLoading: true, error: null }),
    setError: assign({ error: ({ event }) => (event as { type: "LOAD_ERROR"; error: string }).error, isLoading: false }),
    setLoaded: assign({ isLoading: false }),
    updateLocation: assign({
      currentLocation: ({ event }) => (event as { type: "LOCATION_UPDATED"; location: string; cfi: string; progress: number }).location,
      currentCfi: ({ event }) => (event as { type: "LOCATION_UPDATED"; location: string; cfi: string; progress: number }).cfi,
      progress: ({ event }) => (event as { type: "LOCATION_UPDATED"; location: string; cfi: string; progress: number }).progress,
    }),
    updateToc: assign({
      toc: ({ event }) => (event as { type: "TOC_UPDATED"; toc: TocItem[] }).toc,
    }),
  },
}).createMachine({
  id: "reader",
  initial: "idle",
  context: {
    isLoading: false,
    error: null,
    currentLocation: "",
    currentCfi: "",
    progress: 0,
    toc: [],
  },
  states: {
    idle: {
      on: {
        INIT: { target: "loading", actions: "setLoading" },
      },
    },
    loading: {
      on: {
        LOAD_SUCCESS: { target: "ready", actions: "setLoaded" },
        LOAD_ERROR: { target: "failed", actions: "setError" },
        LOCATION_UPDATED: { actions: "updateLocation" },
        TOC_UPDATED: { actions: "updateToc" },
      },
    },
    ready: {
      on: {
        INIT: { target: "loading", actions: "setLoading" },
        LOCATION_UPDATED: { actions: "updateLocation" },
        TOC_UPDATED: { actions: "updateToc" },
        NAVIGATE_PREV: {},
        NAVIGATE_NEXT: {},
        NAVIGATE_TO: {},
      },
    },
    failed: {
      on: {
        INIT: { target: "loading", actions: "setLoading" },
      },
    },
  },
})
