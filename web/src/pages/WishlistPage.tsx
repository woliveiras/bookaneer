import { AuthLayout } from "../components/layout/AppLayout"
import { WishlistList } from "../containers/wishlist/WishlistList"

export function WishlistPage() {
  return (
    <AuthLayout>
      <h2 className="text-2xl font-bold mb-6">Wishlist</h2>
      <WishlistList />
    </AuthLayout>
  )
}
