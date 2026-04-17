import { useQuery } from "@tanstack/react-query"
import { AuthLayout } from "../components/layout/AppLayout"
import { Button } from "../components/ui"
import { useAuthStore } from "../store/auth/auth.store"
import { type ActiveCommand, wantedApi } from "../lib/api"

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

export function SystemPage() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  const status = useQuery<SystemStatus>({
    queryKey: ["status"],
    queryFn: () => fetch("/api/v1/system/status").then((r) => r.json()),
    enabled: isAuthenticated,
  })

  const recentLogs = useQuery<ActiveCommand[]>({
    queryKey: ["commands", "recent"],
    queryFn: () => wantedApi.getRecentCommands(15),
    enabled: isAuthenticated,
    refetchInterval: 5000,
  })

  const getStatusBadge = (cmdStatus: string) => {
    const styles: Record<string, string> = {
      completed: "bg-green-500/10 text-green-500 border-green-500/20",
      failed: "bg-red-500/10 text-red-500 border-red-500/20",
      running: "bg-blue-500/10 text-blue-500 border-blue-500/20",
      queued: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
      cancelled: "bg-gray-500/10 text-gray-500 border-gray-500/20",
    }
    return styles[cmdStatus] || "bg-gray-500/10 text-gray-500"
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
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium border ${getStatusBadge(cmd.status)}`}
                    >
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
                      {cmd.payload &&
                        Object.keys(cmd.payload).length > 0 &&
                        cmd.name !== "DownloadMonitor" && (
                          <p className="text-muted-foreground text-xs mt-1">
                            {cmd.payload.bookTitle
                              ? `Book: ${cmd.payload.bookTitle}`
                              : cmd.payload.bookId
                                ? `Book ID: ${cmd.payload.bookId}`
                                : null}
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
}
