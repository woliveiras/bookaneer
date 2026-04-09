import { AuthLayout } from "../components/layout/AppLayout"
import { BookList } from "../containers/books/BookList"

export function BooksPage() {
  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Books</h2>
      <BookList />
    </AuthLayout>
  )
}
