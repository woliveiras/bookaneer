import { AuthLayout } from "../components/layout/AppLayout"
import { AuthorList } from "../containers/authors/AuthorList"

export function AuthorsPage() {
  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Authors</h2>
      <AuthorList />
    </AuthLayout>
  )
}
