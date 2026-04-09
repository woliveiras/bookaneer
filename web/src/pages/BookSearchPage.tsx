import { Link } from "@tanstack/react-router"
import { AuthLayout } from "../components/layout/AppLayout"
import { BookDetails } from "../containers/search/BookDetails"
import type { MetadataBookResult } from "../lib/api"

interface BookSearchPageProps {
  book: MetadataBookResult
  autoSearch: boolean
  existingBookId?: number
}

export function BookSearchPage({ book, autoSearch, existingBookId }: BookSearchPageProps) {
  if (!book.title) {
    return (
      <AuthLayout>
        <div className="text-center py-12">
          <p className="text-muted-foreground">No book selected</p>
          <Link to="/search" className="text-primary underline mt-2 inline-block">
            Back to Search
          </Link>
        </div>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <BookDetails book={book} autoSearch={autoSearch} existingBookId={existingBookId} />
    </AuthLayout>
  )
}
