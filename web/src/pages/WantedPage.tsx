import { AuthLayout } from "../components/layout/AppLayout"
import { WantedList } from "../containers/wanted/WantedList"

export function WantedPage() {
  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Wishlist</h2>
      <WantedList />
    </AuthLayout>
  )
}
