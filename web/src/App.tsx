import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { AuthProvider, useAuth } from "./contexts/AuthContext"
import { LoginPage } from "./components/auth"
import { AuthorList } from "./components/authors"
import { BookList } from "./components/books"
import { MetadataSearch } from "./components/metadata"
import { SettingsGeneral } from "./components/settings"
import { Reader } from "./components/reader"
import { IndexerList, IndexerOptions, InteractiveSearch } from "./components/indexers"
import { Button } from "./components/ui"

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

type Tab = "library" | "authors" | "books" | "search" | "settings" | "system"

function AppContent() {
  const { isAuthenticated, isLoading: authLoading, logout, user } = useAuth()
  const [activeTab, setActiveTab] = useState<Tab>("library")
  const [readingBookFileId, setReadingBookFileId] = useState<number | null>(null)

  const health = useQuery<HealthResponse>({
    queryKey: ["health"],
    queryFn: () => fetch("/api/v1/system/health").then((r) => r.json()),
    enabled: isAuthenticated,
  })

  const status = useQuery<SystemStatus>({
    queryKey: ["status"],
    queryFn: () => fetch("/api/v1/system/status").then((r) => r.json()),
    enabled: isAuthenticated && activeTab === "system",
  })

  // Show loading spinner while checking auth
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

  // Show login page if not authenticated
  if (!isAuthenticated) {
    return <LoginPage />
  }

  // Show reader in fullscreen mode
  if (readingBookFileId !== null) {
    return (
      <Reader
        bookFileId={readingBookFileId}
        onClose={() => setReadingBookFileId(null)}
      />
    )
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
        <nav className="container mx-auto px-4" aria-label="Main navigation">
          <ul className="flex gap-1 -mb-px" role="tablist">
            {(
              [
                { id: "library", label: "Library" },
                { id: "authors", label: "Authors" },
                { id: "books", label: "Books" },
                { id: "search", label: "Search" },
                { id: "settings", label: "Settings" },
                { id: "system", label: "System" },
              ] as const
            ).map((tab) => (
              <li key={tab.id} role="presentation">
                <Button
                  variant="ghost"
                  role="tab"
                  aria-selected={activeTab === tab.id}
                  aria-controls={`${tab.id}-panel`}
                  className={`rounded-none border-b-2 ${
                    activeTab === tab.id
                      ? "border-primary text-primary"
                      : "border-transparent text-muted-foreground hover:text-foreground"
                  }`}
                  onClick={() => setActiveTab(tab.id)}
                >
                  {tab.label}
                </Button>
              </li>
            ))}
          </ul>
        </nav>
      </header>

      <main className="container mx-auto px-4 py-8">
        {activeTab === "library" && (
          <div id="library-panel" role="tabpanel" aria-labelledby="library-tab">
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="text-lg font-semibold mb-2">Welcome to Bookaneer</h2>
              <p className="text-muted-foreground">
                Your self-hosted ebook collection manager. Connect to indexers, manage your library,
                and read your books anywhere.
              </p>
            </div>
          </div>
        )}

        {activeTab === "authors" && (
          <div id="authors-panel" role="tabpanel" aria-labelledby="authors-tab">
            <h2 className="text-2xl font-bold mb-6">Authors</h2>
            <AuthorList />
          </div>
        )}

        {activeTab === "books" && (
          <div id="books-panel" role="tabpanel" aria-labelledby="books-tab">
            <h2 className="text-2xl font-bold mb-6">Books</h2>
            <BookList onOpenReader={setReadingBookFileId} />
          </div>
        )}

        {activeTab === "search" && (
          <div id="search-panel" role="tabpanel" aria-labelledby="search-tab">
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
          </div>
        )}

        {activeTab === "settings" && (
          <div id="settings-panel" role="tabpanel" aria-labelledby="settings-tab">
            <h2 className="text-2xl font-bold mb-6">Settings</h2>
            <div className="space-y-8">
              <SettingsGeneral />
              <hr className="border-border" />
              <section>
                <h3 className="text-lg font-semibold mb-4">Indexers</h3>
                <IndexerList />
              </section>
              <hr className="border-border" />
              <IndexerOptions />
            </div>
          </div>
        )}

        {activeTab === "system" && (
          <div id="system-panel" role="tabpanel" aria-labelledby="system-tab">
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
          </div>
        )}
      </main>
    </div>
  )
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  )
}

export default App
