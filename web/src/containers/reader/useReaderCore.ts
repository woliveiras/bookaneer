import { useActorRef, useSelector } from "@xstate/react"
import { useCallback, useEffect, useRef } from "react"
import type {
  FoliateView,
  RelocateDetail,
  TocItem,
} from "../../components/reader/readerConfig"
import { useReaderBookFile } from "../../hooks/useReader"
import { readerApi } from "../../lib/api"
import { readerMachine } from "../../features/reader/reader.machine"
import { useReaderSettingsStore } from "../../store/reader/reader-settings.store"
import { useReaderProgress } from "./useReaderProgress"
import { useReaderSettings } from "./useReaderSettings"

/** Keep a stable reference to the latest value of a callback or variable. */
function useLatest<T>(value: T) {
  const ref = useRef(value)
  ref.current = value
  return ref
}

export interface ReaderCoreState {
  containerRef: React.RefObject<HTMLDivElement | null>
  viewRef: React.RefObject<FoliateView | null>
  isLoading: boolean
  error: string | null
  currentLocation: string
  currentCfi: string
  progress: number
  toc: TocItem[]
  handlePrev: () => Promise<void>
  handleNext: () => Promise<void>
  handleTocNavigate: (href: string) => Promise<void>
  handleBookmarkNavigate: (position: string) => Promise<void>
}

export function useReaderCore(bookFileId: number): ReaderCoreState {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<FoliateView | null>(null)

  const actorRef = useActorRef(readerMachine)
  const isLoading = useSelector(actorRef, (s) => s.context.isLoading)
  const error = useSelector(actorRef, (s) => s.context.error)
  const currentLocation = useSelector(actorRef, (s) => s.context.currentLocation)
  const currentCfi = useSelector(actorRef, (s) => s.context.currentCfi)
  const progress = useSelector(actorRef, (s) => s.context.progress)
  const toc = useSelector(actorRef, (s) => s.context.toc)

  const { applyStyles } = useReaderSettings(viewRef)
  const settingsTheme = useReaderSettingsStore((s) => s.theme)
  const { savedProgress, saveProgress } = useReaderProgress(bookFileId)

  const { data: bookFile } = useReaderBookFile(bookFileId)

  // Stable refs to avoid re-triggering the init effect
  const saveProgressRef = useLatest(saveProgress)
  const applyStylesRef = useLatest(applyStyles)
  const savedProgressRef = useLatest(savedProgress)
  const currentCfiRef = useLatest(currentCfi)

  // biome-ignore lint/correctness/useExhaustiveDependencies: useLatest refs are stable containers — their .current is read inside callbacks, not reactive deps. settings.theme triggers intentional reinit.
  useEffect(() => {
    const container = containerRef.current
    if (!container || !bookFile) return

    let cancelled = false

    const initReader = async () => {
      try {
        actorRef.send({ type: "INIT", bookFile, bookFileId })

        const view = document.createElement("foliate-view") as FoliateView
        view.style.width = "100%"
        view.style.height = "100%"
        viewRef.current = view

        view.addEventListener("relocate", ((e: CustomEvent<RelocateDetail>) => {
          if (cancelled) return
          const detail = e.detail
          if (detail.cfi) {
            actorRef.send({
              type: "LOCATION_UPDATED",
              location: detail.tocItem?.label || "",
              cfi: detail.cfi,
              progress: detail.fraction || 0,
            })
            saveProgressRef.current(detail.cfi, detail.fraction || 0)
          }
        }) as EventListener)

        view.addEventListener("load", (() => {
          if (cancelled) return
          if (view.book?.toc) {
            actorRef.send({ type: "TOC_UPDATED", toc: view.book.toc })
          }
          applyStylesRef.current()
        }) as EventListener)

        container.innerHTML = ""
        container.appendChild(view)

        const contentUrl = readerApi.getContentUrl(bookFileId)
        const response = await fetch(contentUrl)
        if (!response.ok) throw new Error("Failed to fetch book content")
        const blob = await response.blob()

        // Wrap blob as File with a name so foliate-js can detect the format
        const ext = bookFile.format || "epub"
        const file = new File([blob], `book.${ext}`, { type: blob.type })

        if (cancelled) return
        await view.open(file)

        // Prefer current in-session location when reinitializing (e.g. theme switch)
        const position = currentCfiRef.current || savedProgressRef.current?.position
        if (position) {
          try {
            await view.goTo(position)
          } catch {
            console.warn("Could not restore position:", position)
          }
        }

        if (!cancelled) actorRef.send({ type: "LOAD_SUCCESS" })
      } catch (err) {
        if (cancelled) return
        console.error("Failed to initialize reader:", err)
        actorRef.send({ type: "LOAD_ERROR", error: err instanceof Error ? err.message : "Failed to load book" })
      }
    }

    initReader()

    return () => {
      cancelled = true
      if (viewRef.current) container.innerHTML = ""
      viewRef.current = null
    }
  }, [bookFile, bookFileId, settingsTheme])

  const handlePrev = useCallback(async () => {
    await viewRef.current?.prev()
  }, [])

  const handleNext = useCallback(async () => {
    await viewRef.current?.next()
  }, [])

  const handleTocNavigate = useCallback(async (href: string) => {
    await viewRef.current?.goTo(href)
  }, [])

  const handleBookmarkNavigate = useCallback(async (position: string) => {
    await viewRef.current?.goTo(position)
  }, [])

  return {
    containerRef,
    viewRef,
    isLoading,
    error,
    currentLocation,
    currentCfi,
    progress,
    toc,
    handlePrev,
    handleNext,
    handleTocNavigate,
    handleBookmarkNavigate,
  }
}
