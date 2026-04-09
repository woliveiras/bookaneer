import { useEffect, useRef, useState, useCallback } from "react"
import {
  useReaderBookFile,
  useReadingProgress,
  useSaveProgress,
  useBookmarks,
  useCreateBookmark,
  useDeleteBookmark,
} from "../../hooks/useReader"
import { readerApi } from "../../lib/api"
import { Button, Badge } from "../ui"
import {
  FORMAT_LABELS,
  THEMES,
  FONTS,
  DEFAULT_SETTINGS,
  loadSettings,
  saveSettings,
  type ThemeKey,
  type ReaderSettings,
  type TocItem,
  type FoliateView,
  type RelocateDetail,
} from "./readerConfig"
import { TocList } from "./TocList"

// Import foliate-js view component
import "foliate-js/view.js"

// Configure PDF.js worker for PDF support
import * as pdfjsLib from "pdfjs-dist"
pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
  "pdfjs-dist/build/pdf.worker.min.mjs",
  import.meta.url
).toString()

interface ReaderProps {
  bookFileId: number
  onClose: () => void
}

export function Reader({ bookFileId, onClose }: ReaderProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<FoliateView | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentLocation, setCurrentLocation] = useState<string>("")
  const [currentCfi, setCurrentCfi] = useState<string>("")
  const [progress, setProgress] = useState(0)

  // UI panels
  const [showSettings, setShowSettings] = useState(false)
  const [showToc, setShowToc] = useState(false)
  const [showBookmarks, setShowBookmarks] = useState(false)
  const [toc, setToc] = useState<TocItem[]>([])

  // Reader settings
  const [settings, setSettings] = useState<ReaderSettings>(loadSettings)

  const { data: bookFile } = useReaderBookFile(bookFileId)
  const { data: savedProgress } = useReadingProgress(bookFileId)
  const saveProgressMutation = useSaveProgress(bookFileId)

  // Bookmarks
  const { data: bookmarks } = useBookmarks(bookFileId)
  const createBookmarkMutation = useCreateBookmark(bookFileId)
  const deleteBookmarkMutation = useDeleteBookmark(bookFileId)

  // Update settings and persist
  const updateSettings = useCallback((updates: Partial<ReaderSettings>) => {
    setSettings((prev) => {
      const next = { ...prev, ...updates }
      saveSettings(next)
      return next
    })
  }, [])

  // Apply styles to foliate-view
  const applyStyles = useCallback(() => {
    const view = viewRef.current
    if (!view?.renderer?.setStyles) return

    const theme = THEMES[settings.theme]
    const css = `
      @import url('https://fonts.googleapis.com/css2?family=Literata:opsz,wght@7..72,400;7..72,700&display=swap');
      
      html {
        background: ${theme.bg} !important;
        color: ${theme.fg} !important;
      }
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

  // Apply styles when settings change
  useEffect(() => {
    applyStyles()
  }, [applyStyles])

  // Debounced save progress
  const saveProgressCallback = useCallback(
    (cfi: string, percentage: number) => {
      saveProgressMutation.mutate({ position: cfi, percentage })
    },
    [saveProgressMutation]
  )

  // Initialize reader
  useEffect(() => {
    const container = containerRef.current
    if (!container || !bookFile) return

    const initReader = async () => {
      try {
        setIsLoading(true)
        setError(null)

        // Create foliate-view element
        const view = document.createElement("foliate-view") as FoliateView
        view.style.width = "100%"
        view.style.height = "100%"
        viewRef.current = view

        // Listen for location changes
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

        // Listen for book loaded
        view.addEventListener("load", (() => {
          // Extract TOC
          if (view.book?.toc) {
            setToc(view.book.toc)
          }
          // Apply initial styles
          applyStyles()
        }) as EventListener)

        // Clear container and append view
        container.innerHTML = ""
        container.appendChild(view)

        // Fetch the book content
        const contentUrl = readerApi.getContentUrl(bookFileId)
        const response = await fetch(contentUrl)
        if (!response.ok) {
          throw new Error("Failed to fetch book content")
        }
        const blob = await response.blob()

        // Open the book
        await view.open(blob)

        // Restore saved position if available
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
      if (viewRef.current) {
        container.innerHTML = ""
      }
      viewRef.current = null
    }
  }, [bookFile, bookFileId, savedProgress?.position, saveProgressCallback, applyStyles])

  // Navigation handlers
  const handlePrev = useCallback(async () => {
    if (viewRef.current) {
      await viewRef.current.prev()
    }
  }, [])

  const handleNext = useCallback(async () => {
    if (viewRef.current) {
      await viewRef.current.next()
    }
  }, [])

  const handleTocClick = useCallback(async (href: string) => {
    if (viewRef.current) {
      await viewRef.current.goTo(href)
      setShowToc(false)
    }
  }, [])

  const handleBookmarkClick = useCallback(async (position: string) => {
    if (viewRef.current) {
      await viewRef.current.goTo(position)
      setShowBookmarks(false)
    }
  }, [])

  const handleAddBookmark = useCallback(() => {
    if (!currentCfi) return
    const title = currentLocation || `Page ${Math.round(progress * 100)}%`
    createBookmarkMutation.mutate({ position: currentCfi, title, note: "" })
  }, [currentCfi, currentLocation, progress, createBookmarkMutation])

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't handle if settings, TOC, or bookmarks is open
      if (showSettings || showToc || showBookmarks) {
        if (e.key === "Escape") {
          setShowSettings(false)
          setShowToc(false)
          setShowBookmarks(false)
        }
        return
      }

      if (e.key === "ArrowLeft" || e.key === "PageUp") {
        e.preventDefault()
        handlePrev()
      } else if (e.key === "ArrowRight" || e.key === "PageDown" || e.key === " ") {
        e.preventDefault()
        handleNext()
      } else if (e.key === "Escape") {
        e.preventDefault()
        onClose()
      } else if (e.key === "t" || e.key === "T") {
        e.preventDefault()
        setShowToc((prev) => !prev)
      } else if (e.key === "s" || e.key === "S") {
        e.preventDefault()
        setShowSettings((prev) => !prev)
      } else if (e.key === "b" || e.key === "B") {
        e.preventDefault()
        setShowBookmarks((prev) => !prev)
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [handlePrev, handleNext, onClose, showSettings, showToc, showBookmarks])

  const theme = THEMES[settings.theme]

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col"
      style={{ background: theme.bg, color: theme.fg }}
    >
      {/* Header */}
      <header
        className="flex items-center justify-between border-b px-4 py-2"
        style={{ borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5" }}
      >
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            aria-label="Close reader"
            style={{ color: theme.fg }}
          >
            ← Back
          </Button>
          <div className="flex items-center gap-2">
            {bookFile?.format && (
              <Badge variant="outline" className="text-xs">
                {FORMAT_LABELS[bookFile.format] || bookFile.format.toUpperCase()}
              </Badge>
            )}
            <div className="text-sm">
              <span className="font-medium">{bookFile?.bookTitle || "Loading..."}</span>
              {bookFile?.authorName && (
                <span style={{ opacity: 0.7 }}> by {bookFile.authorName}</span>
              )}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowToc((prev) => !prev)}
            aria-label="Table of contents"
            aria-pressed={showToc}
            style={{ color: theme.fg }}
          >
            📖 TOC
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowBookmarks((prev) => !prev)}
            aria-label="Bookmarks"
            aria-pressed={showBookmarks}
            style={{ color: theme.fg }}
          >
            🔖 Bookmarks
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowSettings((prev) => !prev)}
            aria-label="Reader settings"
            aria-pressed={showSettings}
            style={{ color: theme.fg }}
          >
            ⚙️ Settings
          </Button>
          <span className="text-sm" style={{ opacity: 0.7 }}>
            {currentLocation && <span className="mr-2">{currentLocation}</span>}
            {Math.round(progress * 100)}%
          </span>
        </div>
      </header>

      {/* Main content area */}
      <div className="relative flex-1 overflow-hidden">
        {/* Loading state */}
        {isLoading && (
          <div
            className="absolute inset-0 flex items-center justify-center"
            style={{ background: theme.bg }}
          >
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-current mx-auto mb-4" />
              <p style={{ opacity: 0.7 }}>Loading book...</p>
            </div>
          </div>
        )}

        {/* Error state */}
        {error && (
          <div
            className="absolute inset-0 flex items-center justify-center"
            style={{ background: theme.bg }}
          >
            <div className="text-center">
              <p className="text-red-500 mb-4">{error}</p>
              <Button variant="outline" onClick={onClose}>
                Close
              </Button>
            </div>
          </div>
        )}

        {/* Reader container */}
        <div ref={containerRef} className="h-full w-full" style={{ background: theme.bg }} />

        {/* TOC Panel */}
        {showToc && (
          <div
            className="absolute left-0 top-0 bottom-0 w-80 overflow-y-auto border-r shadow-lg"
            style={{
              background: theme.bg,
              borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5",
            }}
          >
            <div className="p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="font-bold text-lg">Table of Contents</h2>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowToc(false)}
                  aria-label="Close TOC"
                >
                  ✕
                </Button>
              </div>
              {toc.length > 0 ? (
                <TocList items={toc} onSelect={handleTocClick} theme={theme} />
              ) : (
                <p style={{ opacity: 0.7 }}>No table of contents available</p>
              )}
            </div>
          </div>
        )}

        {/* Settings Panel */}
        {showSettings && (
          <div
            className="absolute right-0 top-0 bottom-0 w-80 overflow-y-auto border-l shadow-lg"
            style={{
              background: theme.bg,
              borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5",
            }}
          >
            <div className="p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="font-bold text-lg">Settings</h2>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowSettings(false)}
                  aria-label="Close settings"
                >
                  ✕
                </Button>
              </div>

              {/* Theme */}
              <div className="mb-6">
                <label className="block text-sm font-medium mb-2">Theme</label>
                <div className="flex gap-2">
                  {(Object.keys(THEMES) as ThemeKey[]).map((key) => (
                    <button
                      key={key}
                      type="button"
                      onClick={() => updateSettings({ theme: key })}
                      className={`flex-1 py-2 px-3 rounded border text-sm ${
                        settings.theme === key ? "ring-2 ring-blue-500" : ""
                      }`}
                      style={{
                        background: THEMES[key].bg,
                        color: THEMES[key].fg,
                        borderColor: settings.theme === "dark" ? "#555" : "#ccc",
                      }}
                    >
                      {THEMES[key].name}
                    </button>
                  ))}
                </div>
              </div>

              {/* Font Size */}
              <div className="mb-6">
                <label className="block text-sm font-medium mb-2">
                  Font Size: {settings.fontSize}%
                </label>
                <input
                  type="range"
                  min="75"
                  max="200"
                  step="5"
                  value={settings.fontSize}
                  onChange={(e) => updateSettings({ fontSize: Number(e.target.value) })}
                  className="w-full"
                />
                <div className="flex justify-between text-xs" style={{ opacity: 0.7 }}>
                  <span>75%</span>
                  <span>200%</span>
                </div>
              </div>

              {/* Font Family */}
              <div className="mb-6">
                <label className="block text-sm font-medium mb-2">Font</label>
                <select
                  value={settings.fontFamily}
                  onChange={(e) => updateSettings({ fontFamily: e.target.value })}
                  className="w-full p-2 rounded border"
                  style={{
                    background: theme.bg,
                    color: theme.fg,
                    borderColor: settings.theme === "dark" ? "#555" : "#ccc",
                  }}
                >
                  {FONTS.map((font) => (
                    <option key={font.value} value={font.value}>
                      {font.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Line Height */}
              <div className="mb-6">
                <label className="block text-sm font-medium mb-2">
                  Line Height: {settings.lineHeight.toFixed(1)}
                </label>
                <input
                  type="range"
                  min="1"
                  max="2.5"
                  step="0.1"
                  value={settings.lineHeight}
                  onChange={(e) => updateSettings({ lineHeight: Number(e.target.value) })}
                  className="w-full"
                />
                <div className="flex justify-between text-xs" style={{ opacity: 0.7 }}>
                  <span>1.0</span>
                  <span>2.5</span>
                </div>
              </div>

              {/* Keyboard shortcuts */}
              <div className="mt-8 pt-4 border-t" style={{ borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5" }}>
                <h3 className="font-medium mb-2">Keyboard Shortcuts</h3>
                <dl className="text-sm space-y-1" style={{ opacity: 0.7 }}>
                  <div className="flex justify-between">
                    <dt>Previous page</dt>
                    <dd>← / PageUp</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt>Next page</dt>
                    <dd>→ / PageDown / Space</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt>Table of contents</dt>
                    <dd>T</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt>Bookmarks</dt>
                    <dd>B</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt>Settings</dt>
                    <dd>S</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt>Close reader</dt>
                    <dd>Esc</dd>
                  </div>
                </dl>
              </div>
            </div>
          </div>
        )}

        {/* Bookmarks Panel */}
        {showBookmarks && (
          <div
            className="absolute right-0 top-0 bottom-0 w-80 overflow-y-auto border-l shadow-lg"
            style={{
              background: theme.bg,
              borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5",
            }}
          >
            <div className="p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="font-bold text-lg">Bookmarks</h2>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowBookmarks(false)}
                  aria-label="Close bookmarks"
                >
                  ✕
                </Button>
              </div>

              {/* Add bookmark button */}
              <Button
                variant="outline"
                size="sm"
                className="w-full mb-4"
                onClick={handleAddBookmark}
                disabled={!currentCfi || createBookmarkMutation.isPending}
                style={{
                  borderColor: settings.theme === "dark" ? "#555" : "#ccc",
                  color: theme.fg,
                }}
              >
                {createBookmarkMutation.isPending ? "Adding..." : "🔖 Add Bookmark Here"}
              </Button>

              {/* Bookmark list */}
              {bookmarks && bookmarks.length > 0 ? (
                <ul className="space-y-2">
                  {bookmarks.map((bookmark) => (
                    <li
                      key={bookmark.id}
                      className="flex items-start justify-between p-2 rounded border"
                      style={{ borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5" }}
                    >
                      <button
                        type="button"
                        onClick={() => handleBookmarkClick(bookmark.position)}
                        className="flex-1 text-left text-sm hover:underline"
                        style={{ color: theme.fg }}
                      >
                        {bookmark.title || "Untitled"}
                        <span className="block text-xs" style={{ opacity: 0.6 }}>
                          {new Date(bookmark.createdAt).toLocaleDateString()}
                        </span>
                      </button>
                      <button
                        type="button"
                        onClick={() => deleteBookmarkMutation.mutate(bookmark.id)}
                        className="ml-2 text-red-500 hover:text-red-700 text-sm"
                        aria-label="Delete bookmark"
                        disabled={deleteBookmarkMutation.isPending}
                      >
                        🗑️
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p style={{ opacity: 0.7 }} className="text-sm">
                  No bookmarks yet. Add one by clicking the button above.
                </p>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Footer navigation */}
      <footer
        className="flex items-center justify-between border-t px-4 py-2"
        style={{ borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5" }}
      >
        <Button
          variant="ghost"
          size="sm"
          onClick={handlePrev}
          aria-label="Previous page"
          style={{ color: theme.fg }}
        >
          ← Previous
        </Button>

        {/* Progress bar */}
        <div className="flex-1 mx-4">
          <div
            className="h-1 rounded-full overflow-hidden"
            style={{ background: settings.theme === "dark" ? "#333" : "#e5e5e5" }}
          >
            <div
              className="h-full transition-all duration-300"
              style={{
                width: `${progress * 100}%`,
                background: settings.theme === "dark" ? "#6ea8fe" : "#0d6efd",
              }}
            />
          </div>
        </div>

        <Button
          variant="ghost"
          size="sm"
          onClick={handleNext}
          aria-label="Next page"
          style={{ color: theme.fg }}
        >
          Next →
        </Button>
      </footer>
    </div>
  )
}


