import { useState } from "react"
import { createRouter, createRootRoute, createRoute, Outlet, Link, useNavigate, useLocation } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { AuthProvider, useAuth } from "./contexts/AuthContext"
import { LoginPage } from "./components/auth"
import { AuthorList } from "./components/authors"
import { BookList } from "./components/books"
import { SettingsGeneral, RootFolderList } from "./components/settings"
import { Reader } from "./components/reader"
import { IndexerList, IndexerOptions } from "./components/indexers"
import { DownloadClientList } from "./components/download"
import { UnifiedSearch } from "./components/search/UnifiedSearch"
import { BookDetails } from "./components/search/BookDetails"
import { WantedList, QueueList, HistoryList, BlocklistList } from "./components/wanted"
import { Button } from "./components/ui"
import { bookApi, authorApi, wantedApi, type MetadataBookResult, type ActiveCommand } from "./lib/api"

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
    { to: "/wanted", label: "Wanted" },
    { to: "/activity", label: "Activity" },
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

// Author detail route
const authorDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/author/$authorId",
  component: function AuthorDetailPage() {
    const { authorId } = authorDetailRoute.useParams()
    const navigate = useNavigate()
    
    const { data: author, isLoading, error } = useQuery({
      queryKey: ["author", authorId],
      queryFn: () => authorApi.get(Number(authorId)),
    })

    if (isLoading) {
      return (
        <AuthLayout>
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        </AuthLayout>
      )
    }

    if (error || !author) {
      return (
        <AuthLayout>
          <div className="text-center py-12">
            <p className="text-destructive mb-4">Failed to load author</p>
            <Button onClick={() => navigate({ to: "/authors" })}>Back to Authors</Button>
          </div>
        </AuthLayout>
      )
    }

    return (
      <AuthLayout>
        <div className="space-y-6">
          <div className="flex items-center gap-4">
            <Button variant="outline" size="sm" onClick={() => navigate({ to: "/authors" })}>
              ← Back
            </Button>
            <h2 className="text-2xl font-bold">{author.name}</h2>
          </div>
          <BookList authorId={Number(authorId)} />
        </div>
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
        <BookDetails book={book} />
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
          <RootFolderSettings />
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

// Root folder settings section - collapsible
function RootFolderSettings() {
  const [isOpen, setIsOpen] = useState(true) // Default open - important for first setup
  return (
    <div className="border rounded-lg">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-full p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <span className="text-lg font-semibold">Root Folders</span>
        <span className="text-muted-foreground">{isOpen ? "▼" : "▶"}</span>
      </button>
      {isOpen && (
        <div className="p-4 border-t">
          <RootFolderList />
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

    const recentLogs = useQuery<ActiveCommand[]>({
      queryKey: ["commands", "recent"],
      queryFn: () => wantedApi.getRecentCommands(15),
      enabled: isAuthenticated,
      refetchInterval: 5000, // Refresh every 5 seconds
    })

    const getStatusBadge = (status: string) => {
      const styles: Record<string, string> = {
        completed: "bg-green-500/10 text-green-500 border-green-500/20",
        failed: "bg-red-500/10 text-red-500 border-red-500/20",
        running: "bg-blue-500/10 text-blue-500 border-blue-500/20",
        queued: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
        cancelled: "bg-gray-500/10 text-gray-500 border-gray-500/20",
      }
      return styles[status] || "bg-gray-500/10 text-gray-500"
    }

    const formatTime = (dateStr: string) => {
      const date = new Date(dateStr)
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })
    }

    const getErrorMessage = (cmd: ActiveCommand): string | null => {
      if (cmd.status !== "failed" || !cmd.result) return null
      const error = cmd.result.error || cmd.result.message
      return typeof error === "string" ? error : null
    }

    return (
      <AuthLayout>
        <h2 className="text-2xl font-bold mb-6">System Status</h2>
        
        {/* System Info */}
        <div className="rounded-lg border border-border bg-card p-6 mb-6">
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

        {/* Recent Logs */}
        <div className="rounded-lg border border-border bg-card">
          <div className="flex items-center justify-between p-4 border-b border-border">
            <h3 className="font-semibold">Recent Tasks</h3>
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">Auto-refresh every 5s</span>
              <Button 
                variant="outline" 
                size="sm" 
                onClick={() => recentLogs.refetch()}
                disabled={recentLogs.isRefetching}
              >
                {recentLogs.isRefetching ? "Refreshing..." : "Refresh"}
              </Button>
            </div>
          </div>
          
          <div className="divide-y divide-border max-h-96 overflow-y-auto">
            {recentLogs.isLoading ? (
              <div className="p-4 text-center text-muted-foreground">Loading logs...</div>
            ) : recentLogs.data?.length === 0 ? (
              <div className="p-4 text-center text-muted-foreground">No recent tasks</div>
            ) : (
              recentLogs.data?.map((cmd) => {
                const errorMsg = getErrorMessage(cmd)
                return (
                  <div key={cmd.id} className="p-3 text-sm hover:bg-muted/50">
                    <div className="flex items-start gap-3">
                      <span className={`px-2 py-0.5 rounded text-xs font-medium border ${getStatusBadge(cmd.status)}`}>
                        {cmd.status}
                      </span>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{cmd.name}</span>
                          <span className="text-muted-foreground text-xs">
                            {formatTime(cmd.startedAt || cmd.queuedAt)}
                          </span>
                        </div>
                        {errorMsg && (
                          <p className="text-red-400 text-xs mt-1 truncate" title={errorMsg}>
                            {errorMsg}
                          </p>
                        )}
                        {cmd.payload && Object.keys(cmd.payload).length > 0 && cmd.name !== "DownloadMonitor" && (
                          <p className="text-muted-foreground text-xs mt-1">
                            {cmd.payload.bookTitle ? `Book: ${cmd.payload.bookTitle}` : 
                             cmd.payload.bookId ? `Book ID: ${cmd.payload.bookId}` : null}
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                )
              })
            )}
          </div>
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

// Library book detail route
const libraryBookDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/book/$bookId",
  component: function LibraryBookDetailPage() {
    const { bookId } = libraryBookDetailRoute.useParams()
    const navigate = useNavigate()
    const queryClient = useQueryClient()
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
    
    const { data: book, isLoading, error } = useQuery({
      queryKey: ["book", bookId],
      queryFn: () => bookApi.get(Number(bookId)),
    })

    const deleteMutation = useMutation({
      mutationFn: () => bookApi.delete(Number(bookId)),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["books"] })
        navigate({ to: "/books" })
      },
    })

    if (isLoading) {
      return (
        <AuthLayout>
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
          </div>
        </AuthLayout>
      )
    }

    if (error || !book) {
      return (
        <AuthLayout>
          <div className="text-center py-12">
            <p className="text-destructive mb-4">Failed to load book</p>
            <Button onClick={() => navigate({ to: "/books" })}>Back to Books</Button>
          </div>
        </AuthLayout>
      )
    }

    const hasFile = book.files && book.files.length > 0

    return (
      <AuthLayout>
        <div className="space-y-6">
          <div className="flex items-center gap-4">
            <Button variant="outline" size="sm" onClick={() => navigate({ to: "/books" })}>
              ← Back
            </Button>
          </div>

          <div className="flex gap-6">
            {/* Cover */}
            <div className="flex-shrink-0 w-32 h-48 bg-muted rounded-lg overflow-hidden">
              {book.imageUrl ? (
                <img src={book.imageUrl} alt={book.title} className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-4xl">📖</div>
              )}
            </div>

            {/* Info */}
            <div className="flex-1 space-y-4">
              <div>
                <h2 className="text-2xl font-bold">{book.title}</h2>
                <p className="text-muted-foreground">
                  by <Link to="/author/$authorId" params={{ authorId: String(book.authorId) }} className="underline hover:text-foreground">{book.authorName}</Link>
                </p>
              </div>

              <div className="flex flex-wrap gap-2">
                <span className={`px-2 py-1 rounded text-xs ${book.monitored ? 'bg-green-500/20 text-green-500' : 'bg-muted text-muted-foreground'}`}>
                  {book.monitored ? 'Monitored' : 'Not Monitored'}
                </span>
                <span className={`px-2 py-1 rounded text-xs ${hasFile ? 'bg-blue-500/20 text-blue-500' : 'bg-yellow-500/20 text-yellow-500'}`}>
                  {hasFile ? 'Downloaded' : 'Missing'}
                </span>
              </div>

              {book.releaseDate && (
                <p className="text-sm text-muted-foreground">
                  Released: {new Date(book.releaseDate).toLocaleDateString()}
                </p>
              )}

              {book.isbn13 && (
                <p className="text-sm text-muted-foreground">ISBN: {book.isbn13}</p>
              )}

              {book.overview && (
                <p className="text-sm text-muted-foreground line-clamp-4">{book.overview}</p>
              )}

              <div className="flex gap-2 pt-4">
                {hasFile && (
                  <Button onClick={() => navigate({ to: "/read/$bookId", params: { bookId: String(book.id) } })}>
                    📖 Read
                  </Button>
                )}
                <Button variant="outline" onClick={() => navigate({ to: "/search/book", search: {
                  title: book.title,
                  authors: book.authorName,
                  foreignId: book.foreignId || undefined,
                  isbn13: book.isbn13 || undefined,
                  coverUrl: book.imageUrl || undefined,
                  publishedYear: book.releaseDate ? String(new Date(book.releaseDate).getFullYear()) : undefined,
                } })}>
                  🔍 Manual Search
                </Button>
                <Button
                  variant="outline"
                  className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
                  onClick={() => setShowDeleteConfirm(true)}
                >
                  🗑️ Delete
                </Button>
              </div>

              {/* Delete confirmation dialog */}
              {showDeleteConfirm && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                  <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4">
                    <h3 className="text-lg font-semibold mb-2">Delete Book?</h3>
                    <p className="text-muted-foreground mb-4">
                      Are you sure you want to delete "{book.title}"? This will remove it from your library.
                    </p>
                    <div className="flex gap-2 justify-end">
                      <Button variant="outline" onClick={() => setShowDeleteConfirm(false)}>
                        Cancel
                      </Button>
                      <Button
                        variant="destructive"
                        onClick={() => deleteMutation.mutate()}
                        disabled={deleteMutation.isPending}
                      >
                        {deleteMutation.isPending ? "Deleting..." : "Delete"}
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </AuthLayout>
    )
  },
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
