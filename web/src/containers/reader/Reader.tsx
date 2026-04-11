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
            style={{ color: theme.fg }}
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
            style={{ color: theme.fg }}
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
            style={{ color: theme.fg }}
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
                  style={{ color: theme.fg }}
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
        style={{ borderColor: settings.theme === "dark" ? "#333" : "#e5e5e5" }}
      >
        <Button
          variant="ghost"
          size="sm"
          onClick={handlePrev}
          aria-label="Previous page"
          style={{ color: theme.fg }}
        >
          <ChevronLeft className="h-4 w-4 mr-1" />
          Previous
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
          Next
          <ChevronRight className="h-4 w-4 ml-1" />
        </Button>
      </footer>
    </div>
  )
}
