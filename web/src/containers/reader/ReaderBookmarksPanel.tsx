import type { UseMutationResult } from "@tanstack/react-query"
import { Bookmark as BookmarkIcon, Trash2, X } from "lucide-react"
import { THEMES } from "../../components/reader/readerConfig"
import { Button } from "../../components/ui"
import type { CreateBookmarkInput } from "../../lib/api"
import type { Bookmark } from "../../lib/types/reader"
import { useReaderSettingsStore } from "../../store/reader/reader-settings.store"

interface ReaderBookmarksPanelProps {
  bookmarks: Bookmark[] | undefined
  currentCfi: string
  currentLocation: string
  progress: number
  onNavigate: (position: string) => void
  onClose: () => void
  createBookmarkMutation: UseMutationResult<unknown, Error, CreateBookmarkInput>
  deleteBookmarkMutation: UseMutationResult<unknown, Error, number>
}

export function ReaderBookmarksPanel({
  bookmarks,
  currentCfi,
  currentLocation,
  progress,
  onNavigate,
  onClose,
  createBookmarkMutation,
  deleteBookmarkMutation,
}: ReaderBookmarksPanelProps) {
  const theme_key = useReaderSettingsStore((s) => s.theme)
  const theme = THEMES[theme_key]
  const borderColor =
    theme_key === "dark" ? "#333" : theme_key === "sepia" ? "#d4c9b0" : "#e5e5e5"

  const handleAddBookmark = () => {
    const title =
      currentLocation ||
      (progress > 0 ? `${Math.round(progress * 100)}% through book` : "Beginning")
    createBookmarkMutation.mutate({ position: currentCfi, title, note: "" })
  }

  return (
    <div
      className="absolute right-0 top-0 bottom-0 w-80 overflow-y-auto border-l shadow-lg"
      style={{ background: theme.bg, borderColor }}
    >
      <div className="p-4">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-bold text-lg">Bookmarks</h2>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            aria-label="Close bookmarks"
            style={{ color: theme.fg }}
          >
            <X className="h-4 w-4" />
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
            borderColor: theme_key === "dark" ? "#555" : "#ccc",
            color: theme.fg,
          }}
        >
          {createBookmarkMutation.isPending ? (
            "Adding..."
          ) : (
            <>
              <BookmarkIcon className="w-4 h-4 inline" /> Add Bookmark Here
            </>
          )}
        </Button>

        {/* Bookmark list */}
        {bookmarks && bookmarks.length > 0 ? (
          <ul className="space-y-2">
            {bookmarks.map((bookmark) => (
              <li
                key={bookmark.id}
                className="flex items-start justify-between p-2 rounded border"
                style={{ borderColor }}
              >
                <button
                  type="button"
                  onClick={() => onNavigate(bookmark.position)}
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
                  <Trash2 className="w-4 h-4" />
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
  )
}
