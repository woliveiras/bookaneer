import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate, useParams } from "@tanstack/react-router"
import { useState } from "react"
import { AuthLayout } from "../components/layout/AppLayout"
import { PageError, PageLoading } from "../components/common"
import { Button } from "../components/ui"
import { bookApi } from "../lib/api"
import { navigateToBookSearch } from "../lib/navigation"
import { useReportWrongContent } from "../hooks/useWanted"
import { AlertTriangle, BookOpen, Search, Trash2 } from "lucide-react"

export function LibraryBookDetailPage() {
  const { bookId } = useParams({ from: "/book/$bookId" })
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [showWrongContentConfirm, setShowWrongContentConfirm] = useState(false)
  const wrongContentMutation = useReportWrongContent()

  const {
    data: book,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["book", bookId],
    queryFn: () => bookApi.get(Number(bookId)),
  })

  const deleteMutation = useMutation({
    mutationFn: () => bookApi.delete(Number(bookId)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["books"] })
      navigate({ to: "/books" })
    },
  })

  if (isLoading) {
    return (
      <AuthLayout>
        <PageLoading />
      </AuthLayout>
    )
  }

  if (error || !book) {
    return (
      <AuthLayout>
        <PageError
          message="Failed to load book"
          onBack={() => navigate({ to: "/books" })}
          backLabel="Back to Books"
        />
      </AuthLayout>
    )
  }

  const hasFile = book.files && book.files.length > 0
  const hasContentMismatch = book.files?.some((f) => f.contentMismatch) ?? false

  return (
    <AuthLayout>
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Button variant="outline" size="sm" onClick={() => navigate({ to: "/books" })}>
            ← Back
          </Button>
        </div>

        <div className="flex gap-6">
          {/* Cover */}
          <div className="shrink-0 w-32 h-48 bg-muted rounded-lg overflow-hidden">
            {book.imageUrl ? (
              <img src={book.imageUrl} alt={book.title} className="w-full h-full object-cover" />
            ) : (
              <div className="w-full h-full flex items-center justify-center"><BookOpen className="w-8 h-8 text-muted-foreground" /></div>
            )}
          </div>

          {/* Info */}
          <div className="flex-1 space-y-4">
            <div>
              <h2 className="text-2xl font-bold">{book.title}</h2>
              <p className="text-muted-foreground">
                by{" "}
                <Link
                  to="/author/$authorId"
                  params={{ authorId: String(book.authorId) }}
                  className="underline hover:text-foreground"
                >
                  {book.authorName}
                </Link>
              </p>
            </div>

            <div className="flex flex-wrap gap-2">
              <span
                className={`px-2 py-1 rounded text-xs ${book.monitored ? "bg-green-500/20 text-green-500" : "bg-muted text-muted-foreground"}`}
              >
                {book.monitored ? "Monitored" : "Not Monitored"}
              </span>
              <span
                className={`px-2 py-1 rounded text-xs ${hasFile ? "bg-blue-500/20 text-blue-500" : "bg-yellow-500/20 text-yellow-500"}`}
              >
                {hasFile ? "Downloaded" : "Missing"}
              </span>
              {book.fileFormat && (
                <span className="px-2 py-1 rounded text-xs bg-muted text-muted-foreground uppercase">
                  {book.fileFormat}
                </span>
              )}
            </div>

            {book.releaseDate && (
              <p className="text-sm text-muted-foreground">
                Released: {new Date(book.releaseDate).toLocaleDateString()}
              </p>
            )}

            {book.isbn13 && <p className="text-sm text-muted-foreground">ISBN: {book.isbn13}</p>}

            {book.overview && (
              <p className="text-sm text-muted-foreground line-clamp-4">{book.overview}</p>
            )}

            {/* Content mismatch warning */}
            {hasContentMismatch && (
              <div className="flex items-start gap-2 p-3 rounded-lg bg-amber-500/10 border border-amber-500/30">
                <AlertTriangle className="w-5 h-5 text-amber-500 shrink-0 mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-amber-500">Possible wrong content</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    The downloaded file's metadata doesn't match this book. Open to verify, or report as wrong content to try another source.
                  </p>
                </div>
              </div>
            )}

            <div className="flex gap-2 pt-4">
              {hasFile && (
                <Button
                  onClick={() =>
                    navigate({
                      to: "/read/$bookId",
                      params: { bookId: String(book.id) },
                    })
                  }
                >
                  <BookOpen className="w-4 h-4" /> Read
                </Button>
              )}
              {hasFile && (
                <Button
                  variant="outline"
                  className="text-amber-500 border-amber-500/50 hover:bg-amber-500/10"
                  onClick={() => setShowWrongContentConfirm(true)}
                >
                  <AlertTriangle className="w-4 h-4" /> Wrong Content
                </Button>
              )}
              <Button
                variant="outline"
                onClick={() => navigateToBookSearch(navigate, book)}
              >
                <Search className="w-4 h-4" /> Manual Search
              </Button>
              <Button
                variant="outline"
                className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
                onClick={() => setShowDeleteConfirm(true)}
              >
                <Trash2 className="w-4 h-4" /> Delete
              </Button>
            </div>

            {/* Delete confirmation dialog */}
            {showDeleteConfirm && (
              <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4">
                  <h3 className="text-lg font-semibold mb-2">Delete Book?</h3>
                  <p className="text-muted-foreground mb-4">
                    Are you sure you want to delete "{book.title}"? This will remove it from your
                    library.
                  </p>
                  <div className="flex gap-2 justify-end">
                    <Button variant="outline" onClick={() => setShowDeleteConfirm(false)}>
                      Cancel
                    </Button>
                    <Button
                      variant="destructive"
                      onClick={() => deleteMutation.mutate()}
                      disabled={deleteMutation.isPending}
                    >
                      {deleteMutation.isPending ? "Deleting..." : "Delete"}
                    </Button>
                  </div>
                </div>
              </div>
            )}

            {/* Wrong content confirmation dialog */}
            {showWrongContentConfirm && (
              <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4">
                  <h3 className="text-lg font-semibold mb-2">Report Wrong Content?</h3>
                  <p className="text-muted-foreground mb-4">
                    This will remove the current file, blocklist this source, and automatically search for an alternative download.
                  </p>
                  <div className="flex gap-2 justify-end">
                    <Button variant="outline" onClick={() => setShowWrongContentConfirm(false)}>
                      Cancel
                    </Button>
                    <Button
                      className="bg-amber-500 hover:bg-amber-600 text-white"
                      onClick={() => {
                        wrongContentMutation.mutate(book.id, {
                          onSuccess: () => setShowWrongContentConfirm(false),
                        })
                      }}
                      disabled={wrongContentMutation.isPending}
                    >
                      {wrongContentMutation.isPending ? "Reporting..." : "Report & Retry"}
                    </Button>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </AuthLayout>
  )
}
