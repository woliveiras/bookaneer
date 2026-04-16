import { AuthLayout } from "../components/layout/AppLayout"
import { Badge } from "../components/ui/Badge"
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/Card"
import { useAuthors } from "../hooks/useAuthors"
import { useBooks } from "../hooks/useBooks"
import { useRootFolders } from "../hooks/useRootFolders"
import { useDownloadQueue, useHistory } from "../hooks/useWanted"
import { useWishlist } from "../hooks/useWishlist"
import { formatBytes } from "../lib/format"

function StatCard({
  label,
  value,
  detail,
}: {
  label: string
  value: string | number
  detail?: string
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <p className="text-sm text-muted-foreground">{label}</p>
        <p className="text-2xl sm:text-3xl font-bold mt-1">{value}</p>
        {detail && <p className="text-xs text-muted-foreground mt-1">{detail}</p>}
      </CardContent>
    </Card>
  )
}

function eventLabel(eventType: string): string {
  switch (eventType) {
    case "grabbed":
      return "Grabbed"
    case "bookImported":
      return "Imported"
    case "downloadCompleted":
      return "Downloaded"
    case "downloadFailed":
      return "Failed"
    case "bookFileDeleted":
      return "Deleted"
    case "bookFileRenamed":
      return "Renamed"
    case "contentMismatch":
      return "Mismatch"
    case "wrongContent":
      return "Wrong Content"
    case "metadataExtracted":
      return "Metadata"
    default:
      return eventType
  }
}

function eventBadgeVariant(eventType: string): "default" | "secondary" | "destructive" | "outline" {
  switch (eventType) {
    case "bookImported":
    case "downloadCompleted":
      return "default"
    case "downloadFailed":
    case "contentMismatch":
    case "wrongContent":
      return "destructive"
    default:
      return "secondary"
  }
}

export function LibraryPage() {
  const { data: booksData } = useBooks({ limit: 1 })
  const { data: authorsData } = useAuthors({ limit: 1 })
  const { data: wishlistData } = useWishlist()
  const { data: queue } = useDownloadQueue()
  const { data: rootFolders } = useRootFolders()
  const { data: history } = useHistory({ limit: 10 })

  const totalBooks = booksData?.totalRecords ?? 0
  const totalAuthors = authorsData?.totalRecords ?? 0
  const inWishlist = wishlistData?.totalRecords ?? 0
  const booksOnDisk = totalBooks - inWishlist

  const activeDownloads = queue?.filter((q) => q.status === "downloading") ?? []
  const completedQueue = queue?.filter((q) => q.status === "completed") ?? []
  const failedQueue = queue?.filter((q) => q.status === "failed") ?? []

  const diskTotal = rootFolders?.reduce((acc, f) => acc + (f.totalSpace ?? 0), 0) ?? 0

  return (
    <AuthLayout>
      <div className="space-y-6">
        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <StatCard label="Total Books" value={totalBooks} />
          <StatCard label="Authors" value={totalAuthors} />
          <StatCard
            label="On Disk"
            value={booksOnDisk >= 0 ? booksOnDisk : 0}
            detail={totalBooks > 0 ? `${booksOnDisk} of ${totalBooks} books have files` : undefined}
          />
          <StatCard label="In Wishlist" value={inWishlist} detail="Not yet downloaded" />
        </div>

        {/* Disk usage */}
        {rootFolders && rootFolders.length > 0 && diskTotal > 0 && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Disk Usage</CardTitle>
            </CardHeader>
            <CardContent className="pb-4">
              {rootFolders.map((folder) => (
                <div key={folder.id} className="mb-3 last:mb-0">
                  <div className="flex justify-between text-sm mb-1">
                    <span className="font-medium">{folder.name}</span>
                    <span className="text-muted-foreground">
                      {formatBytes(folder.freeSpace ?? 0)} free of{" "}
                      {formatBytes(folder.totalSpace ?? 0)}
                    </span>
                  </div>
                  <div className="h-2 bg-muted rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary rounded-full transition-all"
                      style={{
                        width: `${folder.totalSpace ? ((folder.totalSpace - (folder.freeSpace ?? 0)) / folder.totalSpace) * 100 : 0}%`,
                      }}
                    />
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Active downloads */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Download Queue</CardTitle>
            </CardHeader>
            <CardContent>
              {!queue?.length ? (
                <p className="text-sm text-muted-foreground py-2">No items in queue</p>
              ) : (
                <div className="space-y-3">
                  {activeDownloads.map((item) => (
                    <div key={item.id}>
                      <div className="flex justify-between text-sm">
                        <span className="truncate font-medium">{item.bookTitle || item.title}</span>
                        <span className="text-muted-foreground ml-2 shrink-0">
                          {Math.round(item.progress * 100)}%
                        </span>
                      </div>
                      <div className="h-1.5 bg-muted rounded-full overflow-hidden mt-1">
                        <div
                          className="h-full bg-primary rounded-full transition-all"
                          style={{ width: `${item.progress * 100}%` }}
                        />
                      </div>
                    </div>
                  ))}
                  {completedQueue.length > 0 && (
                    <div className="flex justify-between text-sm text-muted-foreground">
                      <span>{completedQueue.length} completed</span>
                    </div>
                  )}
                  {failedQueue.length > 0 && (
                    <div className="flex justify-between text-sm text-destructive">
                      <span>{failedQueue.length} failed</span>
                    </div>
                  )}
                  {activeDownloads.length === 0 &&
                    completedQueue.length === 0 &&
                    failedQueue.length === 0 && (
                      <p className="text-sm text-muted-foreground">Queue is idle</p>
                    )}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Recent activity */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Recent Activity</CardTitle>
            </CardHeader>
            <CardContent>
              {!history?.length ? (
                <p className="text-sm text-muted-foreground py-2">No recent activity</p>
              ) : (
                <div className="space-y-2">
                  {history.map((item) => (
                    <div key={item.id} className="flex items-center justify-between gap-2 text-sm">
                      <div className="flex items-center gap-2 min-w-0">
                        <Badge
                          variant={eventBadgeVariant(item.eventType)}
                          className="text-xs shrink-0"
                        >
                          {eventLabel(item.eventType)}
                        </Badge>
                        <span className="truncate">{item.bookTitle || item.sourceTitle}</span>
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0">
                        {new Date(item.date).toLocaleDateString()}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </AuthLayout>
  )
}
