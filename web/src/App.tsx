import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { AuthorList } from "./components/authors"
import { BookList } from "./components/books"
import { MetadataSearch } from "./components/metadata"
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

type Tab = "library" | "authors" | "books" | "search" | "system"

function App() {
  const [activeTab, setActiveTab] = useState<Tab>("library")

  const health = useQuery<HealthResponse>({
    queryKey: ["health"],
    queryFn: () => fetch("/api/v1/system/health").then((r) => r.json()),
  })

  const status = useQuery<SystemStatus>({
    queryKey: ["status"],
    queryFn: () => fetch("/api/v1/system/status").then((r) => r.json()),
    enabled: activeTab === "system",
  })

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
            <BookList />
          </div>
        )}

        {activeTab === "search" && (
          <div id="search-panel" role="tabpanel" aria-labelledby="search-tab">
            <h2 className="text-2xl font-bold mb-6">Search Metadata</h2>
            <p className="text-muted-foreground mb-6">
              Search for authors and books across OpenLibrary, Google Books, and Hardcover.
            </p>
            <MetadataSearch />
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

export default App
