import { useState } from "react"
import { createRouter, createRootRoute, createRoute, Outlet, Link, useNavigate } from "@tanstack/react-router"
import { useQuery } from "@tanstack/react-query"
import { AuthProvider, useAuth } from "./contexts/AuthContext"
import { LoginPage } from "./components/auth"
import { AuthorList } from "./components/authors"
import { BookList } from "./components/books"
import { MetadataSearch } from "./components/metadata"
import { SettingsGeneral } from "./components/settings"
import { Reader } from "./components/reader"
import { IndexerList, IndexerOptions, InteractiveSearch } from "./components/indexers"
import { DownloadClientList } from "./components/download"
import { Button } from "./components/ui"
import { bookApi } from "./lib/api"

// Types
interface HealthResponse {
  status: string
}

interface SystemStatus {
  version: string
  buildTime: string
  osName: string
  osArch: string
  runtimeVersion: string
  startTime: string
  appDataDir: string
  libraryDir: string
}

// Auth-protected layout wrapper
function AuthLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading: authLoading, logout, user } = useAuth()

  const health = useQuery<HealthResponse>({
    queryKey: ["health"],
    queryFn: () => fetch("/api/v1/system/health").then((r) => r.json()),
    enabled: isAuthenticated,
  })

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <LoginPage />
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-2xl font-bold text-foreground">📚 Bookaneer</h1>
          <div className="flex items-center gap-4">
            {health.isLoading ? (
              <span className="text-muted-foreground">Checking...</span>
            ) : health.data?.status === "ok" ? (
              <span className="inline-flex items-center gap-1 text-green-600">
                <span className="h-2 w-2 rounded-full bg-green-500" />
                Connected
              </span>
            ) : (
              <span className="text-destructive">Disconnected</span>
            )}
            {user && (
              <span className="text-sm text-muted-foreground">
                {user.username || "API Key"}
              </span>
            )}
            <Button variant="outline" size="sm" onClick={logout}>
              Sign Out
            </Button>
          </div>
        </div>
        <Navigation />
      </header>

      <main className="container mx-auto px-4 py-8">
        {children}
      </main>
    </div>
  )
}

// Navigation component using TanStack Router Link
function Navigation() {
  const navItems = [
    { to: "/", label: "Library" },
    { to: "/authors", label: "Authors" },
    { to: "/books", label: "Books" },
    { to: "/search", label: "Search" },
    { to: "/settings", label: "Settings" },
    { to: "/system", label: "System" },
  ] as const

  return (
    <nav className="container mx-auto px-4" aria-label="Main navigation">
      <ul className="flex gap-1 -mb-px" role="tablist">
        {navItems.map((item) => (
          <li key={item.to} role="presentation">
            <Link
              to={item.to}
              className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium rounded-none border-b-2 transition-colors"
              activeProps={{
                className: "border-primary text-primary",
              }}
              inactiveProps={{
                className: "border-transparent text-muted-foreground hover:text-foreground",
              }}
              activeOptions={{ exact: item.to === "/" }}
            >
              {item.label}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  )
}

// Root route
const rootRoute = createRootRoute({
  component: () => (
    <AuthProvider>
      <Outlet />
    </AuthProvider>
  ),
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
        <div className="space-y-8">
          <section>
            <h3 className="text-lg font-semibold mb-4">Search Indexers</h3>
            <p className="text-muted-foreground mb-4">
              Search for ebooks across your configured Newznab/Torznab indexers.
            </p>
            <InteractiveSearch />
          </section>
          <hr className="border-border" />
          <section>
            <h3 className="text-lg font-semibold mb-4">Search Metadata</h3>
            <p className="text-muted-foreground mb-4">
              Search for authors and books across OpenLibrary, Google Books, and Hardcover.
            </p>
            <MetadataSearch />
          </section>
        </div>
      </AuthLayout>
    )
  },
})

// Settings route
const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings",
  component: function SettingsPage() {
    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">Settings</h2>
        <div className="space-y-4">
          <GeneralSettings />
          <IndexerSettings />
          <DownloadClientSettings />
        </div>
      </AuthLayout>
    )
  },
})

// General settings section - collapsible
function GeneralSettings() {
  const [isOpen, setIsOpen] = useState(true) // Default open
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">General</span>
        <span className="text-muted-foreground">{isOpen ? "▼" : "▶"}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t">
          <SettingsGeneral />
        </div>
      )}
    </div>
  )
}

// Indexer settings section - collapsible
function IndexerSettings() {
  const [isOpen, setIsOpen] = useState(false)
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">Indexers</span>
        <span className="text-muted-foreground">{isOpen ? "▼" : "▶"}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t space-y-6">
          <IndexerList />
          <hr className="border-border" />
          <IndexerOptions />
        </div>
      )}
    </div>
  )
}

// Download client settings section - collapsible
function DownloadClientSettings() {
  const [isOpen, setIsOpen] = useState(false)
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">Download Clients</span>
        <span className="text-muted-foreground">{isOpen ? "▼" : "▶"}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t">
          <DownloadClientList />
        </div>
      )}
    </div>
  )
}

// System route
const systemRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/system",
  component: function SystemPage() {
    const { isAuthenticated } = useAuth()
    
    const status = useQuery<SystemStatus>({
      queryKey: ["status"],
      queryFn: () => fetch("/api/v1/system/status").then((r) => r.json()),
      enabled: isAuthenticated,
    })

    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">System Status</h2>
        <div className="rounded-lg border border-border bg-card p-6">
          {status.isLoading ? (
            <p className="text-muted-foreground">Loading...</p>
          ) : status.data ? (
            <dl className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <dt className="text-muted-foreground">Version</dt>
                <dd className="font-mono">{status.data.version}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Runtime</dt>
                <dd className="font-mono">{status.data.runtimeVersion}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Platform</dt>
                <dd className="font-mono">
                  {status.data.osName}/{status.data.osArch}
                </dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Started</dt>
                <dd className="font-mono">{new Date(status.data.startTime).toLocaleString()}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Data Dir</dt>
                <dd className="font-mono">{status.data.appDataDir}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Library Dir</dt>
                <dd className="font-mono">{status.data.libraryDir}</dd>
              </div>
            </dl>
          ) : (
            <p className="text-destructive">Failed to load status</p>
          )}
        </div>
      </AuthLayout>
    )
  },
})

// Reader route with dynamic bookId - fetches book details to get file
const readerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/read/$bookId",
  component: function ReaderPage() {
    const { bookId } = readerRoute.useParams()
    const navigate = useNavigate()

    const { data: book, isLoading, error } = useQuery({
      queryKey: ["book", bookId],
      queryFn: () => bookApi.get(Number(bookId)),
    })

    if (isLoading) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-background">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
            <p className="text-muted-foreground">Loading book...</p>
          </div>
        </div>
      )
    }

    if (error || !book) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-background">
          <div className="text-center">
            <p className="text-destructive mb-4">Failed to load book</p>
            <Button onClick={() => navigate({ to: "/books" })}>Back to Books</Button>
          </div>
        </div>
      )
    }

    if (!book.files || book.files.length === 0) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-background">
          <div className="text-center">
            <p className="text-muted-foreground mb-4">No file available for this book</p>
            <Button onClick={() => navigate({ to: "/books" })}>Back to Books</Button>
          </div>
        </div>
      )
    }

    return (
      <Reader
        bookFileId={book.files[0].id}
        onClose={() => navigate({ to: "/books" })}
      />
    )
  },
})

// Route tree
const routeTree = rootRoute.addChildren([
  indexRoute,
  authorsRoute,
  booksRoute,
  searchRoute,
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
