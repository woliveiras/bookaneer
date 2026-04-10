import { createRootRoute, createRoute, createRouter, useNavigate } from "@tanstack/react-router"
import { RootLayout } from "./components/layout/AppLayout"
import type { MetadataBookResult } from "./lib/api"
import { ActivityPage } from "./pages/ActivityPage"
import { AuthorDetailPage } from "./pages/AuthorDetailPage"
import { AuthorsPage } from "./pages/AuthorsPage"
import { LibraryBookDetailPage } from "./pages/LibraryBookDetailPage"
import { ReaderPage } from "./pages/ReaderPage"
import { BookSearchPage } from "./pages/BookSearchPage"
import { BooksPage } from "./pages/BooksPage"
import { LibraryPage } from "./pages/LibraryPage"
import { SearchPage } from "./pages/SearchPage"
import { SettingsPage } from "./pages/SettingsPage"
import { SystemPage } from "./pages/SystemPage"
import { WantedPage } from "./pages/WantedPage"

// Root route
const rootRoute = createRootRoute({
  component: RootLayout,
})

// Library (Home) route
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: LibraryPage,
})

// Authors route
const authorsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/authors",
  component: AuthorsPage,
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
  component: BooksPage,
})

// Search route
const searchRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/search",
  component: SearchPage,
})

// Wanted route (missing books)
const wantedRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/wanted",
  component: WantedPage,
})

// Activity route (download queue)
const activityRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/activity",
  validateSearch: (search: Record<string, unknown>) => ({
    tab: (search.tab as string) || "queue",
  }),
  component: function ActivityPageRoute() {
    const navigate = useNavigate()
    const { tab } = activityRoute.useSearch()

    return (
      <ActivityPage
        tab={tab}
        onTabChange={(newTab) => navigate({ to: "/activity", search: { tab: newTab } })}
      />
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
  component: function BookDetailsPageRoute() {
    const search = bookDetailsRoute.useSearch()

    // Reconstruct book data from search params
    const book: MetadataBookResult = {
      title: search.title,
      authors: search.authors ? search.authors.split("|||") : undefined,
      provider: search.provider || "unknown",
      foreignId: search.foreignId || "",
      publishedYear: search.publishedYear ? parseInt(search.publishedYear, 10) : undefined,
      coverUrl: search.coverUrl,
      isbn13: search.isbn13,
    }

    return (
      <BookSearchPage
        book={book}
        autoSearch={search.autoSearch === "true"}
        existingBookId={search.bookId ? parseInt(search.bookId, 10) : undefined}
      />
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
