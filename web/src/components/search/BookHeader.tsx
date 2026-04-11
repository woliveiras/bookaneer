import { Link, useNavigate } from "@tanstack/react-router"
import { AlertTriangle, ClipboardList, Flag, Library, Lightbulb, Plus, Search } from "lucide-react"
import type { MetadataBookResult } from "../../lib/api"
import { Badge, Button } from "../ui"

interface BookHeaderProps {
  book: MetadataBookResult
  addedToLibrary: boolean
  addingToLibrary: boolean
  addError: string | null
  grabError: string | null
  searchStarted: boolean
  hasRootFolder: boolean
  onAddToLibrary: () => void
  onStartSearch: () => void
}

export function BookHeader({
  book,
  addedToLibrary,
  addingToLibrary,
  addError,
  grabError,
  searchStarted,
  hasRootFolder,
  onAddToLibrary,
  onStartSearch,
}: BookHeaderProps) {
  const navigate = useNavigate()

  return (
    <>
      {/* Book Header */}
      <div className="flex gap-6 p-6 rounded-lg border bg-card">
        {book.coverUrl ? (
          <img
            src={book.coverUrl}
            alt={book.title}
            className="w-32 h-48 object-cover rounded shadow-lg"
          />
        ) : (
          <div className="w-32 h-48 bg-muted rounded flex items-center justify-center">
            <Library className="w-8 h-8 text-muted-foreground" />
          </div>
        )}
        <div className="flex-1">
          <h1 className="text-2xl font-bold">{book.title}</h1>
          {book.authors && book.authors.length > 0 && (
            <p className="text-lg text-muted-foreground mt-1">{book.authors.join(", ")}</p>
          )}
          {book.publishedYear && <p className="text-muted-foreground">{book.publishedYear}</p>}
          <div className="flex flex-wrap gap-2 mt-4">
            <Badge variant="outline">{book.provider}</Badge>
            {book.isbn13 && <Badge variant="secondary">{book.isbn13}</Badge>}
          </div>
          <div className="flex flex-wrap gap-2 mt-4">
            <Button variant="outline" size="sm" onClick={() => navigate({ to: "/search" })}>
              ← Back to Search
            </Button>
            {!addedToLibrary ? (
              <Button
                size="sm"
                variant="default"
                onClick={onAddToLibrary}
                disabled={addingToLibrary}
              >
                {addingToLibrary ? "Adding..." : <><Plus className="w-4 h-4" /> Add to Library</>}
              </Button>
            ) : (
              <Button
                size="sm"
                variant="outline"
                className="text-green-600 border-green-600"
                disabled
              >
                ✓ Added to Library
              </Button>
            )}
            {!searchStarted && (
              <Button size="sm" variant="secondary" onClick={onStartSearch}>
                <Search className="w-4 h-4" /> Manual Search
              </Button>
            )}
          </div>
          {addError && <p className="text-sm text-destructive mt-2">{addError}</p>}
          {grabError && <p className="text-sm text-destructive mt-2">{grabError}</p>}
        </div>
      </div>

      {/* Initial state - search not started */}
      {!searchStarted && (
        <div className="p-6 rounded-lg border bg-card">
          <div className="grid md:grid-cols-2 gap-8">
            {/* Wanted Section */}
            <div className="text-center md:text-left space-y-3">
              <div className="flex items-center gap-2 justify-center md:justify-start">
                <span className="text-2xl"><ClipboardList className="w-6 h-6" /></span>
                <h3 className="text-lg font-semibold">Add to Wanted</h3>
              </div>
              <p className="text-sm text-muted-foreground">
                Click <strong>"Add to Library"</strong> above to monitor this book. Bookaneer will
                automatically search for it when new releases become available.
              </p>
              <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/50 p-2 rounded">
                <span><Lightbulb className="w-4 h-4" /></span>
                <span>
                  Wanted books appear in <strong>Activity → Wanted</strong> tab
                </span>
              </div>
            </div>

            {/* Manual Search Section */}
            <div className="text-center md:text-left space-y-3 md:border-l md:pl-8">
              <div className="flex items-center gap-2 justify-center md:justify-start">
                <span className="text-2xl"><Search className="w-6 h-6" /></span>
                <h3 className="text-lg font-semibold">Manual Search</h3>
              </div>
              <p className="text-sm text-muted-foreground">
                Search digital libraries and torrent indexers right now for "{book.title}". Review
                results and grab what you want.
              </p>
              {!hasRootFolder && (
                <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-md p-3">
                  <p className="text-sm text-yellow-600 dark:text-yellow-400 font-medium">
                    <AlertTriangle className="w-4 h-4" /> No Root Folder
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Configure a root folder in{" "}
                    <Link to="/settings" className="text-primary hover:underline">
                      Settings
                    </Link>{" "}
                    before downloading.
                  </p>
                </div>
              )}
              <Button size="lg" onClick={onStartSearch} className="w-full md:w-auto">
                <Flag className="w-4 h-4" /> Start Manual Search
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
