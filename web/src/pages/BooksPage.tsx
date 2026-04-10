import { AuthLayout } from "../components/layout/AppLayout"
import { BookList } from "../containers/books/BookList"

/**
 * List page pattern: wrap a feature container in AuthLayout.
 * New list pages should follow this same minimal structure.
 */
export function BooksPage() {
  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Books</h2>
      <BookList />
    </AuthLayout>
  )
}
