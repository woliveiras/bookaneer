import { useCallback, useEffect, useRef, useState } from "react"
import {
  type FoliateView,
  loadSettings,
  type ReaderSettings,
  type RelocateDetail,
  saveSettings,
  THEMES,
  type TocItem,
} from "../../components/reader/readerConfig"
import { useReaderBookFile, useReadingProgress, useSaveProgress } from "../../hooks/useReader"
import { readerApi } from "../../lib/api"

export interface ReaderCoreState {
  containerRef: React.RefObject<HTMLDivElement | null>
  viewRef: React.RefObject<FoliateView | null>
  isLoading: boolean
  error: string | null
  currentLocation: string
  currentCfi: string
  progress: number
  toc: TocItem[]
  settings: ReaderSettings
  updateSettings: (updates: Partial<ReaderSettings>) => void
  handlePrev: () => Promise<void>
  handleNext: () => Promise<void>
  handleTocNavigate: (href: string) => Promise<void>
  handleBookmarkNavigate: (position: string) => Promise<void>
}

export function useReaderCore(bookFileId: number): ReaderCoreState {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<FoliateView | null>(null)

  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentLocation, setCurrentLocation] = useState("")
  const [currentCfi, setCurrentCfi] = useState("")
  const [progress, setProgress] = useState(0)
  const [toc, setToc] = useState<TocItem[]>([])
  const [settings, setSettings] = useState<ReaderSettings>(loadSettings)

  const { data: bookFile } = useReaderBookFile(bookFileId)
  const { data: savedProgress } = useReadingProgress(bookFileId)
  const saveProgressMutation = useSaveProgress(bookFileId)

  const updateSettings = useCallback((updates: Partial<ReaderSettings>) => {
    setSettings((prev) => {
      const next = { ...prev, ...updates }
      saveSettings(next)
      return next
    })
  }, [])

  const applyStyles = useCallback(() => {
    const view = viewRef.current
    if (!view?.renderer?.setStyles) return
    const theme = THEMES[settings.theme]
    const css = `
      @import url('https://fonts.googleapis.com/css2?family=Literata:opsz,wght@7..72,400;7..72,700&display=swap');
      html { background: ${theme.bg} !important; color: ${theme.fg} !important; }
      body {
        font-family: ${settings.fontFamily} !important;
        font-size: ${settings.fontSize}% !important;
        line-height: ${settings.lineHeight} !important;
        background: ${theme.bg} !important;
        color: ${theme.fg} !important;
      }
      a { color: ${settings.theme === "dark" ? "#6ea8fe" : "#0d6efd"}; }
    `
    view.renderer.setStyles(css)
  }, [settings])

  useEffect(() => {
    applyStyles()
  }, [applyStyles])

  const saveProgressCallback = useCallback(
    (cfi: string, percentage: number) => {
      saveProgressMutation.mutate({ position: cfi, percentage })
    },
    [saveProgressMutation],
  )

  useEffect(() => {
    const container = containerRef.current
    if (!container || !bookFile) return

    const initReader = async () => {
      try {
        setIsLoading(true)
        setError(null)

        const view = document.createElement("foliate-view") as FoliateView
        view.style.width = "100%"
        view.style.height = "100%"
        viewRef.current = view

        view.addEventListener("relocate", ((e: CustomEvent<RelocateDetail>) => {
          const detail = e.detail
          if (detail.cfi) {
            setCurrentLocation(detail.tocItem?.label || "")
            setCurrentCfi(detail.cfi)
            const percentage = detail.fraction || 0
            setProgress(percentage)
            saveProgressCallback(detail.cfi, percentage)
          }
        }) as EventListener)

        view.addEventListener("load", (() => {
          if (view.book?.toc) {
            setToc(view.book.toc)
          }
          applyStyles()
        }) as EventListener)

        container.innerHTML = ""
        container.appendChild(view)

        const contentUrl = readerApi.getContentUrl(bookFileId)
        const response = await fetch(contentUrl)
        if (!response.ok) throw new Error("Failed to fetch book content")
        const blob = await response.blob()

        await view.open(blob)

        if (savedProgress?.position) {
          try {
            await view.goTo(savedProgress.position)
          } catch {
            console.warn("Could not restore position:", savedProgress.position)
          }
        }

        setIsLoading(false)
      } catch (err) {
        console.error("Failed to initialize reader:", err)
        setError(err instanceof Error ? err.message : "Failed to load book")
        setIsLoading(false)
      }
    }

    initReader()

    return () => {
      if (viewRef.current) container.innerHTML = ""
      viewRef.current = null
    }
  }, [bookFile, bookFileId, savedProgress?.position, saveProgressCallback, applyStyles])

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
    settings,
    updateSettings,
    handlePrev,
    handleNext,
    handleTocNavigate,
    handleBookmarkNavigate,
  }
}
