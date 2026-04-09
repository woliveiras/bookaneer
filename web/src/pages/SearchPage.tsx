import { AuthLayout } from "../components/layout/AppLayout"
import { UnifiedSearch } from "../containers/search/UnifiedSearch"

export function SearchPage() {
  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Search</h2>
      <UnifiedSearch />
    </AuthLayout>
  )
}
