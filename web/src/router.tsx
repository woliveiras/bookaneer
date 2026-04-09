import { createRouter, createRootRoute, createRoute, Link, useNavigate } from "@tanstack/react-router"
import { AuthLayout, RootLayout } from "./components/layout/AppLayout"
import { AuthorList } from "./components/authors"
import { BookList } from "./components/books"
import { UnifiedSearch } from "./components/search/UnifiedSearch"
import { BookDetails } from "./components/search/BookDetails"
import { WantedList, QueueList, HistoryList, BlocklistList } from "./components/wanted"
import { type MetadataBookResult } from "./lib/api"
import { AuthorDetailPage } from "./pages/AuthorDetailPage"
import { SettingsPage } from "./pages/SettingsPage"
import { SystemPage } from "./pages/SystemPage"
import { ReaderPage, LibraryBookDetailPage } from "./pages/BookPages"

// Root route
const rootRoute = createRootRoute({
  component: RootLayout,
})

// Library (Home) route
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: function LibraryPage() {
    return (
      <AuthLayout>
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold mb-2">Welcome to Bookaneer</h2>
          <p className="text-muted-foreground">
            Your self-hosted ebook collection manager. Connect to indexers, manage your library,
            and read your books anywhere.
          </p>
        </div>
      </AuthLayout>
    )
  },
})

// Authors route
const authorsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/authors",
  component: function AuthorsPage() {
    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">Authors</h2>
        <AuthorList />
      </AuthLayout>
    )
  },
})

// Author detail route
const authorDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/author/$authorId",
  component: AuthorDetailPage,
})

// Books route
const booksRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/books",
  component: function BooksPage() {
    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">Books</h2>
        <BookList />
      </AuthLayout>
    )
  },
})

// Search route
const searchRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/search",
  component: function SearchPage() {
    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">Search</h2>
        <UnifiedSearch />
      </AuthLayout>
    )
  },
})

// Wanted route (missing books)
const wantedRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/wanted",
  component: function WantedPage() {
    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">Wanted</h2>
        <WantedList />
      </AuthLayout>
    )
  },
})

// Activity route (download queue)
const activityRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/activity",
  validateSearch: (search: Record<string, unknown>) => ({
    tab: (search.tab as string) || "queue",
  }),
  component: function ActivityPage() {
    const navigate = useNavigate()
    const { tab } = activityRoute.useSearch()
    
    const tabs = [
      { id: "queue", label: "Queue" },
      { id: "history", label: "History" },
      { id: "blocklist", label: "Blocklist" },
    ]

    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">Activity</h2>
        
        {/* Tab navigation */}
        <div className="border-b border-border mb-6">
          <nav className="-mb-px flex space-x-8" aria-label="Activity tabs">
            {tabs.map((t) => (
              <button
                key={t.id}
                onClick={() => navigate({ to: "/activity", search: { tab: t.id } })}
                className={`
                  whitespace-nowrap border-b-2 py-3 px-1 text-sm font-medium transition-colors
                  ${tab === t.id
                    ? "border-primary text-primary"
                    : "border-transparent text-muted-foreground hover:border-muted-foreground/30 hover:text-foreground"
                  }
                `}
                aria-current={tab === t.id ? "page" : undefined}
              >
                {t.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab content */}
        {tab === "queue" && <QueueList />}
        {tab === "history" && <HistoryList />}
        {tab === "blocklist" && <BlocklistList />}
      </AuthLayout>
    )
  },
})

// Book details search params type
interface BookDetailsSearch {
  title: string
  authors?: string
  provider?: string
  foreignId?: string
  publishedYear?: string
  coverUrl?: string
  isbn13?: string
  autoSearch?: string
  bookId?: string
}

// Book details route (child of search)
const bookDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/search/book",
  validateSearch: (search: Record<string, unknown>): BookDetailsSearch => ({
    title: (search.title as string) || "",
    authors: search.authors as string | undefined,
    provider: search.provider as string | undefined,
    foreignId: search.foreignId as string | undefined,
    publishedYear: search.publishedYear as string | undefined,
    coverUrl: search.coverUrl as string | undefined,
    isbn13: search.isbn13 as string | undefined,
    autoSearch: search.autoSearch as string | undefined,
    bookId: search.bookId as string | undefined,
  }),
  component: function BookDetailsPage() {
    const search = bookDetailsRoute.useSearch()
    
    // Reconstruct book data from search params
    const book: MetadataBookResult = {
      title: search.title,
      authors: search.authors ? search.authors.split("|||") : undefined,
      provider: search.provider || "unknown",
      foreignId: search.foreignId || "",
      publishedYear: search.publishedYear ? parseInt(search.publishedYear) : undefined,
      coverUrl: search.coverUrl,
      isbn13: search.isbn13,
    }

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
        <BookDetails 
          book={book} 
          autoSearch={search.autoSearch === "true"} 
          existingBookId={search.bookId ? parseInt(search.bookId) : undefined}
        />
      </AuthLayout>
    )
  },
})

// Settings route
const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings",
  component: SettingsPage,
})

// System route
const systemRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/system",
  component: SystemPage,
})

// Reader route with dynamic bookId
const readerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/read/$bookId",
  component: ReaderPage,
})

// Library book detail route
const libraryBookDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/book/$bookId",
  component: LibraryBookDetailPage,
})

// Route tree
const routeTree = rootRoute.addChildren([
  indexRoute,
  authorsRoute,
  authorDetailRoute,
  booksRoute,
  libraryBookDetailRoute,
  wantedRoute,
  activityRoute,
  searchRoute,
  bookDetailsRoute,
  settingsRoute,
  systemRoute,
  readerRoute,
])

// Create router
export const router = createRouter({ routeTree })

// Type declaration for router
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
