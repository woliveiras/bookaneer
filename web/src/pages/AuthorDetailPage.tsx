import { useQuery } from "@tanstack/react-query"
import { useNavigate, useParams } from "@tanstack/react-router"
import { useState } from "react"
import { AuthLayout } from "../components/layout/AppLayout"
import { PageError, PageLoading } from "../components/common"
import { Button } from "../components/ui"
import { BookList } from "../containers/books/BookList"
import { useDeleteAuthor } from "../hooks/useAuthors"
import { authorApi } from "../lib/api"

export function AuthorDetailPage() {
  const { authorId } = useParams({ from: "/author/$authorId" })
  const navigate = useNavigate()
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [deleteFiles, setDeleteFiles] = useState(true)
  const deleteAuthor = useDeleteAuthor()

  const {
    data: author,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["author", authorId],
    queryFn: () => authorApi.get(Number(authorId)),
  })

  const handleDelete = () => {
    deleteAuthor.mutate(
      { id: Number(authorId), deleteFiles },
      {
        onSuccess: () => {
          navigate({ to: "/authors" })
        },
      },
    )
  }

  if (isLoading) {
    return (
      <AuthLayout>
        <PageLoading />
      </AuthLayout>
    )
  }

  if (error || !author) {
    return (
      <AuthLayout>
        <PageError
          message="Failed to load author"
          onBack={() => navigate({ to: "/authors" })}
          backLabel="Back to Authors"
        />
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="outline" size="sm" onClick={() => navigate({ to: "/authors" })}>
              ← Back
            </Button>
            <h2 className="text-2xl font-bold">{author.name}</h2>
          </div>
          <Button variant="destructive" size="sm" onClick={() => setShowDeleteModal(true)}>
            Delete Author
          </Button>
        </div>
        <BookList authorId={Number(authorId)} />
      </div>

      {/* Delete Confirmation Modal */}
      {showDeleteModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <button
            type="button"
            className="absolute inset-0 bg-black/50"
            onClick={() => setShowDeleteModal(false)}
            onKeyDown={(e) => e.key === "Escape" && setShowDeleteModal(false)}
            aria-label="Close modal"
          />
          <div className="relative bg-background border border-border rounded-lg p-6 max-w-md w-full mx-4 shadow-lg">
            <h3 className="text-lg font-semibold mb-4">Delete Author</h3>
            <p className="text-muted-foreground mb-4">
              Are you sure you want to delete <strong>{author.name}</strong>? This action cannot be
              undone.
            </p>

            <label className="flex items-center gap-2 mb-6 cursor-pointer">
              <input
                type="checkbox"
                checked={deleteFiles}
                onChange={(e) => setDeleteFiles(e.target.checked)}
                className="w-4 h-4 rounded border-border"
              />
              <span className="text-sm">Delete all downloaded files for this author</span>
            </label>

            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setShowDeleteModal(false)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteAuthor.isPending}
              >
                {deleteAuthor.isPending ? "Deleting..." : "Delete"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </AuthLayout>
  )
}
