import { useQuery } from "@tanstack/react-query"

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

function App() {
  const health = useQuery<HealthResponse>({
    queryKey: ["health"],
    queryFn: () => fetch("/api/v1/system/health").then((r) => r.json()),
  })

  const status = useQuery<SystemStatus>({
    queryKey: ["status"],
    queryFn: () => fetch("/api/v1/system/status").then((r) => r.json()),
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
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="grid gap-6">
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold mb-4">System Status</h2>
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

          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold mb-2">Welcome to Bookaneer</h2>
            <p className="text-muted-foreground">
              Your self-hosted ebook collection manager. Connect to indexers, manage your library,
              and read your books anywhere.
            </p>
          </div>
        </div>
      </main>
    </div>
  )
}

export default App
