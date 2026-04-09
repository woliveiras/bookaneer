import { AuthLayout } from "../components/layout/AppLayout"

export function LibraryPage() {
  return (
    <AuthLayout>
      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-2">Welcome to Bookaneer</h2>
        <p className="text-muted-foreground">
          Your self-hosted ebook collection manager. Connect to indexers, manage your library, and
          read your books anywhere.
        </p>
      </div>
    </AuthLayout>
  )
}
