import { useState } from "react"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate, useParams, Link } from "@tanstack/react-router"
import { AuthLayout } from "../components/layout/AppLayout"
import { Reader } from "../components/reader"
import { Button } from "../components/ui"
import { bookApi } from "../lib/api"

export function ReaderPage() {
  const { bookId } = useParams({ from: "/read/$bookId" })
  const navigate = useNavigate()

  const { data: book, isLoading, error } = useQuery({
    queryKey: ["book", bookId],
    queryFn: () => bookApi.get(Number(bookId)),
  })

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
          <p className="text-muted-foreground">Loading book...</p>
        </div>
      </div>
    )
  }

  if (error || !book) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <p className="text-destructive mb-4">Failed to load book</p>
          <Button onClick={() => navigate({ to: "/books" })}>Back to Books</Button>
        </div>
      </div>
    )
  }

  if (!book.files || book.files.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <p className="text-muted-foreground mb-4">No file available for this book</p>
          <Button onClick={() => navigate({ to: "/books" })}>Back to Books</Button>
        </div>
      </div>
    )
  }

  return (
    <Reader
      bookFileId={book.files[0].id}
      onClose={() => navigate({ to: "/books" })}
    />
  )
}

export function LibraryBookDetailPage() {
  const { bookId } = useParams({ from: "/book/$bookId" })
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  const { data: book, isLoading, error } = useQuery({
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
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </AuthLayout>
    )
  }

  if (error || !book) {
    return (
      <AuthLayout>
        <div className="text-center py-12">
          <p className="text-destructive mb-4">Failed to load book</p>
          <Button onClick={() => navigate({ to: "/books" })}>Back to Books</Button>
        </div>
      </AuthLayout>
    )
  }

  const hasFile = book.files && book.files.length > 0

  return (
    <AuthLayout>
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate({ to: "/books" })}
          >
            ← Back
          </Button>
        </div>

        <div className="flex gap-6">
          {/* Cover */}
          <div className="shrink-0 w-32 h-48 bg-muted rounded-lg overflow-hidden">
            {book.imageUrl ? (
              <img
                src={book.imageUrl}
                alt={book.title}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-4xl">
                📖
              </div>
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
            </div>

            {book.releaseDate && (
              <p className="text-sm text-muted-foreground">
                Released: {new Date(book.releaseDate).toLocaleDateString()}
              </p>
            )}

            {book.isbn13 && (
              <p className="text-sm text-muted-foreground">
                ISBN: {book.isbn13}
              </p>
            )}

            {book.overview && (
              <p className="text-sm text-muted-foreground line-clamp-4">
                {book.overview}
              </p>
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
                  📖 Read
                </Button>
              )}
              <Button
                variant="outline"
                onClick={() =>
                  navigate({
                    to: "/search/book",
                    search: {
                      title: book.title,
                      authors: book.authorName,
                      foreignId: book.foreignId || undefined,
                      isbn13: book.isbn13 || undefined,
                      coverUrl: book.imageUrl || undefined,
                      publishedYear: book.releaseDate
                        ? String(new Date(book.releaseDate).getFullYear())
                        : undefined,
                      autoSearch: "true",
                      bookId: String(book.id),
                    },
                  })
                }
              >
                🔍 Manual Search
              </Button>
              <Button
                variant="outline"
                className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
                onClick={() => setShowDeleteConfirm(true)}
              >
                🗑️ Delete
              </Button>
            </div>

            {/* Delete confirmation dialog */}
            {showDeleteConfirm && (
              <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4">
                  <h3 className="text-lg font-semibold mb-2">Delete Book?</h3>
                  <p className="text-muted-foreground mb-4">
                    Are you sure you want to delete "{book.title}"? This will
                    remove it from your library.
                  </p>
                  <div className="flex gap-2 justify-end">
                    <Button
                      variant="outline"
                      onClick={() => setShowDeleteConfirm(false)}
                    >
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
          </div>
        </div>
      </div>
    </AuthLayout>
  );
}
