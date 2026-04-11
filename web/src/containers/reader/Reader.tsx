import { BookOpen, Bookmark, ChevronLeft, ChevronRight, Settings, X } from "lucide-react"
import { useState } from "react"
import { FORMAT_LABELS, THEMES } from "../../components/reader/readerConfig"
import { TocList } from "../../components/reader/TocList"
import { Badge, Button } from "../../components/ui"
import { useBookmarks, useCreateBookmark, useDeleteBookmark, useReaderBookFile } from "../../hooks/useReader"
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
    settings,
    updateSettings,
    handlePrev,
    handleNext,
    handleTocNavigate,
    handleBookmarkNavigate,
  } = useReaderCore(bookFileId)

  const { data: bookFile } = useReaderBookFile(bookFileId)
  const { data: bookmarks } = useBookmarks(bookFileId)
  const createBookmarkMutation = useCreateBookmark(bookFileId)
  const deleteBookmarkMutation = useDeleteBookmark(bookFileId)

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

  const theme = THEMES[settings.theme]

  // Compute theme-aware hover colors
  const hoverBg =
    settings.theme === "dark"
      ? "rgba(255,255,255,0.1)"
      : settings.theme === "sepia"
        ? "rgba(91,70,54,0.1)"
        : "rgba(0,0,0,0.06)"
  const borderColor = settings.theme === "dark" ? "#333" : settings.theme === "sepia" ? "#d4c9b0" : "#e5e5e5"
  const progressTrack = settings.theme === "dark" ? "#333" : settings.theme === "sepia" ? "#d4c9b0" : "#e5e5e5"
  const progressBar = settings.theme === "dark" ? "#6ea8fe" : settings.theme === "sepia" ? "#8b7355" : "#0d6efd"

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
        className="flex items-center justify-between border-b px-4 py-2"
        style={{ borderColor }}
      >
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="sm"
            className="reader-btn"
            onClick={onClose}
            aria-label="Close reader"
          >
            <ChevronLeft className="h-4 w-4 mr-1" />
            Back
          </Button>
          <div className="flex items-center gap-2">
            {bookFile?.format && (
              <Badge
                variant="outline"
                className="text-xs"
                style={{ color: theme.fg, borderColor: theme.fg }}
              >
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
            className="reader-btn"
          >
            <BookOpen className="h-4 w-4 mr-1" />
            TOC
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowBookmarks((prev) => !prev)}
            aria-label="Bookmarks"
            aria-pressed={showBookmarks}
            className="reader-btn"
          >
            <Bookmark className="h-4 w-4 mr-1" />
            Bookmarks
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowSettings((prev) => !prev)}
            aria-label="Reader settings"
            aria-pressed={showSettings}
            className="reader-btn"
          >
            <Settings className="h-4 w-4 mr-1" />
            Settings
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
            className="absolute left-0 top-0 bottom-0 w-80 overflow-y-auto border-r shadow-lg"
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
                  className="reader-btn"
                >
                  <X className="h-4 w-4" />
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
            settings={settings}
            onUpdateSettings={updateSettings}
            onClose={() => setShowSettings(false)}
          />
        )}

        {/* Bookmarks Panel */}
        {showBookmarks && (
          <ReaderBookmarksPanel
            settings={settings}
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
        className="flex items-center justify-between border-t px-4 py-2"
        style={{ borderColor }}
      >
        <Button
          variant="ghost"
          size="sm"
          className="reader-btn"
          onClick={handlePrev}
          aria-label="Previous page"
        >
          <ChevronLeft className="h-4 w-4 mr-1" />
          Previous
        </Button>

        {/* Progress bar */}
        <div className="flex-1 mx-4">
          <div
            className="h-1 rounded-full overflow-hidden"
            style={{ background: progressTrack }}
          >
            <div
              className="h-full transition-all duration-300"
              style={{
                width: `${progress * 100}%`,
                background: progressBar,
              }}
            />
          </div>
        </div>

        <Button
          variant="ghost"
          size="sm"
          className="reader-btn"
          onClick={handleNext}
          aria-label="Next page"
        >
          Next
          <ChevronRight className="h-4 w-4 ml-1" />
        </Button>
      </footer>
    </div>
  )
}
