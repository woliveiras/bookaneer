import { Bookmark, BookOpen, ChevronLeft, ChevronRight, Settings, X } from "lucide-react"
import { useRef, useState } from "react"
import { FORMAT_LABELS, THEMES } from "../../components/reader/readerConfig"
import { TocList } from "../../components/reader/TocList"
import { Badge, Button } from "../../components/ui"
import {
  useBookmarks,
  useCreateBookmark,
  useDeleteBookmark,
  useReaderBookFile,
} from "../../hooks/useReader"
import { useReaderSettingsStore } from "../../store/reader/reader-settings.store"
import { ReaderBookmarksPanel } from "./ReaderBookmarksPanel"
import { ReaderSettingsPanel } from "./ReaderSettingsPanel"
import { useReaderCore } from "./useReaderCore"
import { useReaderKeyboard } from "./useReaderKeyboard"

// Import foliate-js view component
import "foliate-js/view.js"

// Configure PDF.js worker for PDF support
import * as pdfjsLib from "pdfjs-dist"

pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
  "pdfjs-dist/build/pdf.worker.min.mjs",
  import.meta.url,
).toString()

interface ReaderProps {
  bookFileId: number
  onClose: () => void
}

export function Reader({ bookFileId, onClose }: ReaderProps) {
  const [showSettings, setShowSettings] = useState(false)
  const [showToc, setShowToc] = useState(false)
  const [showBookmarks, setShowBookmarks] = useState(false)

  const {
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
  } = useReaderCore(bookFileId)

  const theme_key = useReaderSettingsStore((s) => s.theme)

  const { data: bookFile } = useReaderBookFile(bookFileId)
  const { data: bookmarks } = useBookmarks(bookFileId)
  const createBookmarkMutation = useCreateBookmark(bookFileId)
  const deleteBookmarkMutation = useDeleteBookmark(bookFileId)

  // Touch swipe state
  const touchStartX = useRef<number | null>(null)
  const touchStartY = useRef<number | null>(null)

  const handleTocClick = async (href: string) => {
    await handleTocNavigate(href)
    setShowToc(false)
  }

  const handleBookmarkClick = async (position: string) => {
    await handleBookmarkNavigate(position)
    setShowBookmarks(false)
  }

  useReaderKeyboard({
    viewRef,
    readerReady: !isLoading,
    showSettings,
    showToc,
    showBookmarks,
    setShowSettings,
    setShowToc,
    setShowBookmarks,
    onPrev: handlePrev,
    onNext: handleNext,
    onClose,
  })

  const theme = THEMES[theme_key]

  // Compute theme-aware hover colors
  const hoverBg =
    theme_key === "dark"
      ? "rgba(255,255,255,0.1)"
      : theme_key === "sepia"
        ? "rgba(91,70,54,0.1)"
        : "rgba(0,0,0,0.06)"
  const borderColor =
    theme_key === "dark" ? "#333" : theme_key === "sepia" ? "#d4c9b0" : "#e5e5e5"
  const progressTrack =
    theme_key === "dark" ? "#333" : theme_key === "sepia" ? "#d4c9b0" : "#e5e5e5"
  const progressBar =
    theme_key === "dark" ? "#6ea8fe" : theme_key === "sepia" ? "#8b7355" : "#0d6efd"

  // Handle swipe gestures for page turns
  const handleTouchStart = (e: React.TouchEvent) => {
    touchStartX.current = e.touches[0].clientX
    touchStartY.current = e.touches[0].clientY
  }

  const handleTouchEnd = (e: React.TouchEvent) => {
    if (touchStartX.current === null || touchStartY.current === null) return
    const dx = e.changedTouches[0].clientX - touchStartX.current
    const dy = e.changedTouches[0].clientY - touchStartY.current
    // Only trigger on horizontal swipes with enough distance, and not more vertical than horizontal
    if (Math.abs(dx) > 50 && Math.abs(dx) > Math.abs(dy) * 1.5) {
      if (dx < 0) {
        handleNext()
      } else {
        handlePrev()
      }
    }
    touchStartX.current = null
    touchStartY.current = null
  }

  return (
    <div
      className="reader-root fixed inset-0 z-50 flex flex-col"
      style={{ background: theme.bg, color: theme.fg }}
    >
      {/* Scoped styles for reader buttons — override Tailwind hover with theme colors */}
      <style>{`
        .reader-btn {
          color: ${theme.fg} !important;
          transition: background 150ms ease;
        }
        .reader-btn:hover {
          background: ${hoverBg} !important;
          color: ${theme.fg} !important;
        }
        .reader-btn,
        .reader-btn:hover {
          cursor: pointer;
        }
        .reader-root foliate-view {
          background: ${theme.bg} !important;
        }
        .reader-root foliate-view::part(container),
        .reader-root foliate-view::part(head),
        .reader-root foliate-view::part(foot),
        .reader-root foliate-view::part(filter) {
          background: ${theme.bg} !important;
        }
      `}</style>

      {/* Header */}
      <header
        className="flex items-center justify-between border-b px-2 sm:px-4 py-1 sm:py-2 gap-2"
        style={{ borderColor }}
      >
        <div className="flex items-center gap-2 sm:gap-4 min-w-0">
          <Button
            variant="ghost"
            size="sm"
            className="reader-btn shrink-0 min-h-[44px] min-w-[44px]"
            onClick={onClose}
            aria-label="Close reader"
          >
            <ChevronLeft className="h-5 w-5" aria-hidden="true" />
            <span className="hidden sm:inline ml-1">Back</span>
          </Button>
          <div className="flex items-center gap-2 min-w-0">
            {bookFile?.format && (
              <Badge
                variant="outline"
                className="text-xs shrink-0"
                style={{ color: theme.fg, borderColor: theme.fg }}
              >
                {FORMAT_LABELS[bookFile.format] || bookFile.format.toUpperCase()}
              </Badge>
            )}
            <div className="text-sm min-w-0">
              <span className="font-medium truncate block">
                {bookFile?.bookTitle || "Loading..."}
              </span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1 sm:gap-2 shrink-0">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowToc((prev) => !prev)}
            aria-label="Table of contents"
            aria-pressed={showToc}
            className="reader-btn min-h-[44px] min-w-[44px]"
          >
            <BookOpen className="h-5 w-5" aria-hidden="true" />
            <span className="hidden sm:inline ml-1">TOC</span>
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowBookmarks((prev) => !prev)}
            aria-label="Bookmarks"
            aria-pressed={showBookmarks}
            className="reader-btn min-h-[44px] min-w-[44px]"
          >
            <Bookmark className="h-5 w-5" aria-hidden="true" />
            <span className="hidden sm:inline ml-1">Bookmarks</span>
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowSettings((prev) => !prev)}
            aria-label="Reader settings"
            aria-pressed={showSettings}
            className="reader-btn min-h-[44px] min-w-[44px]"
          >
            <Settings className="h-5 w-5" aria-hidden="true" />
            <span className="hidden sm:inline ml-1">Settings</span>
          </Button>
          <span className="text-xs sm:text-sm shrink-0" style={{ opacity: 0.7 }}>
            {Math.round(progress * 100)}%
          </span>
        </div>
      </header>

      {/* Main content area */}
      <div
        className="relative flex-1 overflow-hidden"
        onTouchStart={handleTouchStart}
        onTouchEnd={handleTouchEnd}
      >
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
            <div className="text-center max-w-md px-4">
              <p className="mb-2 text-lg font-medium" style={{ color: theme.fg }}>
                Unable to open book
              </p>
              <p className="mb-6 text-sm" style={{ color: theme.fg, opacity: 0.7 }}>
                The book file could not be loaded. It may be corrupted or in an unsupported format.
              </p>
              <Button
                variant="outline"
                onClick={onClose}
                style={{ color: theme.fg, borderColor: theme.fg }}
              >
                Back to Library
              </Button>
            </div>
          </div>
        )}

        {/* Reader container */}
        <div ref={containerRef} className="h-full w-full" style={{ background: theme.bg }} />

        {/* TOC Panel */}
        {showToc && (
          <div
            className="absolute left-0 top-0 bottom-0 w-full sm:w-80 overflow-y-auto border-r shadow-lg"
            style={{
              background: theme.bg,
              borderColor,
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
                  className="reader-btn min-h-[44px] min-w-[44px]"
                >
                  <X className="h-5 w-5" aria-hidden="true" />
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
          <ReaderSettingsPanel
            onClose={() => setShowSettings(false)}
          />
        )}

        {/* Bookmarks Panel */}
        {showBookmarks && (
          <ReaderBookmarksPanel
            bookmarks={bookmarks}
            currentCfi={currentCfi}
            currentLocation={currentLocation}
            progress={progress}
            onNavigate={handleBookmarkClick}
            onClose={() => setShowBookmarks(false)}
            createBookmarkMutation={createBookmarkMutation}
            deleteBookmarkMutation={deleteBookmarkMutation}
          />
        )}
      </div>

      {/* Footer navigation */}
      <footer
        className="flex items-center justify-between border-t px-2 sm:px-4 py-1 sm:py-2"
        style={{ borderColor }}
      >
        <Button
          variant="ghost"
          size="sm"
          className="reader-btn min-h-[44px] min-w-[44px]"
          onClick={handlePrev}
          aria-label="Previous page"
        >
          <ChevronLeft className="h-5 w-5" aria-hidden="true" />
          <span className="hidden sm:inline ml-1">Previous</span>
        </Button>

        {/* Progress bar */}
        <div className="flex-1 mx-2 sm:mx-4">
          <div
            className="h-1.5 rounded-full overflow-hidden"
            style={{ background: progressTrack }}
            role="progressbar"
            aria-valuenow={Math.round(progress * 100)}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label="Reading progress"
          >
            <div
              className="h-full transition-all duration-300"
              style={{
                width: `${progress * 100}%`,
                background: progressBar,
              }}
            />
          </div>
          {currentLocation && (
            <p className="text-xs text-center mt-1" style={{ opacity: 0.6 }}>
              {currentLocation}
            </p>
          )}
        </div>

        <Button
          variant="ghost"
          size="sm"
          className="reader-btn min-h-[44px] min-w-[44px]"
          onClick={handleNext}
          aria-label="Next page"
        >
          <span className="hidden sm:inline mr-1">Next</span>
          <ChevronRight className="h-5 w-5" aria-hidden="true" />
        </Button>
      </footer>
    </div>
  )
}
