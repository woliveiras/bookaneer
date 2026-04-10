import { useQuery } from "@tanstack/react-query"
import { useNavigate, useParams } from "@tanstack/react-router"
import { Button } from "../components/ui"
import { Reader } from "../containers/reader/Reader"
import { bookApi } from "../lib/api"

export function ReaderPage() {
  const { bookId } = useParams({ from: "/read/$bookId" })
  const navigate = useNavigate()

  const {
    data: book,
    isLoading,
    error,
  } = useQuery({
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

  return <Reader bookFileId={book.files[0].id} onClose={() => navigate({ to: "/books" })} />
}
