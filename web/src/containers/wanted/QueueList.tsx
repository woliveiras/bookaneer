import { useQueryClient } from "@tanstack/react-query"
import { Inbox } from "lucide-react"
import { useEffect, useState } from "react"
import { Button, Card, CardContent } from "../../components/ui"
import { QueueItemCard, SearchCommandCard } from "../../components/wanted/QueueCards"
import { getDismissedCommands, saveDismissedCommands } from "../../components/wanted/queueHelpers"
import {
  useActiveCommands,
  useDownloadQueue,
  useRecentCommands,
  useRemoveFromQueue,
  useRetryDownload,
} from "../../hooks/useWanted"
import type { QueueItem } from "../../lib/api"
import { queueApi } from "../../lib/api"

export function QueueList() {
  const { data: queue, isLoading, error, refetch } = useDownloadQueue()
  const { data: activeCommands } = useActiveCommands()
  const { data: recentCommands } = useRecentCommands(50)
  const removeMutation = useRemoveFromQueue()
  const retryMutation = useRetryDownload()
  const queryClient = useQueryClient()
  const [itemToRemove, setItemToRemove] = useState<QueueItem | null>(null)
  const [dismissedCommands, setDismissedCommands] = useState<Set<string>>(getDismissedCommands)
  const [isClearing, setIsClearing] = useState(false)

  // Sync dismissed commands to localStorage
  useEffect(() => {
    saveDismissedCommands(dismissedCommands)
  }, [dismissedCommands])

  const handleRemove = (item: QueueItem) => {
    setItemToRemove(item)
  }

  const confirmRemove = async () => {
    if (!itemToRemove) return
    try {
      await removeMutation.mutateAsync(itemToRemove.id)
      setItemToRemove(null)
    } catch (err) {
      console.error("Failed to remove from queue:", err)
    }
  }

  const dismissCommand = (commandId: string) => {
    setDismissedCommands((prev) => new Set([...prev, commandId]))
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    )
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <p className="text-destructive">Failed to load download queue</p>
          <Button variant="outline" onClick={() => refetch()} className="mt-4">
            Try Again
          </Button>
        </CardContent>
      </Card>
    )
  }

  const downloadItems = queue || []

  // Filter book-related commands that aren't dismissed
  const bookCommands = (recentCommands || []).filter(
    (cmd) => ["DownloadGrab"].includes(cmd.name) && !dismissedCommands.has(cmd.id),
  )

  // Active commands (still running)
  const activeBookCommands = (activeCommands || []).filter((cmd) =>
    ["DownloadGrab"].includes(cmd.name),
  )

  // Completed/failed commands — auto-dismiss "Found" if the book already has a completed download
  const completedBookIds = new Set(
    downloadItems.filter((item) => item.status === "completed").map((item) => item.bookId),
  )
  const finishedCommands = bookCommands.filter((cmd) => {
    if (cmd.status !== "completed" && cmd.status !== "failed") return false
    // Auto-dismiss "Found" commands whose book already completed downloading
    if (cmd.status === "completed" && cmd.payload?.bookId) {
      const bookId = cmd.payload.bookId as number
      if (completedBookIds.has(bookId)) return false
    }
    return true
  })

  const failedDownloads = downloadItems.filter((item) => item.status === "failed")
  const completedDownloads = downloadItems.filter((item) => item.status === "completed")

  const handleClearAllFailed = async () => {
    setIsClearing(true)
    try {
      await Promise.allSettled(failedDownloads.map((item) => queueApi.remove(item.id)))
    } finally {
      setIsClearing(false)
      queryClient.invalidateQueries({ queryKey: ["queue"] })
    }
  }

  const clearAllFinishedCommands = async () => {
    setIsClearing(true)
    finishedCommands.forEach((cmd) => {
      dismissCommand(cmd.id)
    })
    try {
      await Promise.allSettled(completedDownloads.map((item) => queueApi.remove(item.id)))
    } finally {
      setIsClearing(false)
      queryClient.invalidateQueries({ queryKey: ["queue"] })
    }
  }

  const hasActiveItems = activeBookCommands.length > 0 || downloadItems.length > 0
  const hasFinishedItems = finishedCommands.length > 0 || completedDownloads.length > 0
  const isEmpty = !hasActiveItems && !hasFinishedItems

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-muted-foreground">
            {activeBookCommands.length > 0 && <span>{activeBookCommands.length} searching</span>}
            {downloadItems.length > 0 && (
              <span className="ml-2">{downloadItems.length} downloading</span>
            )}
            {hasFinishedItems && (
              <span className="ml-2">
                {finishedCommands.length} {finishedCommands.length === 1 ? "result" : "results"}
              </span>
            )}
            {isEmpty && "No activity"}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {finishedCommands.length + completedDownloads.length > 1 && (
            <Button
              variant="outline"
              size="sm"
              onClick={clearAllFinishedCommands}
              disabled={isClearing}
            >
              Clear All Results
            </Button>
          )}
          {failedDownloads.length > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleClearAllFailed}
              disabled={isClearing}
              className="text-destructive border-destructive hover:bg-destructive/10"
            >
              Clear Failed Downloads
            </Button>
          )}
          <Button variant="outline" size="sm" onClick={() => refetch()} disabled={isLoading}>
            Refresh
          </Button>
        </div>
      </div>

      {/* Empty state */}
      {isEmpty && (
        <Card>
          <CardContent className="p-12 text-center">
            <div className="flex justify-center mb-4">
              <Inbox className="w-8 h-8 text-muted-foreground" />
            </div>
            <h3 className="text-lg font-semibold mb-2">No activity</h3>
            <p className="text-muted-foreground">
              Search for books to start downloading. Activity will appear here.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Active Searches */}
      {activeBookCommands.length > 0 && (
        <div className="space-y-2">
          {activeBookCommands.map((cmd) => (
            <SearchCommandCard
              key={cmd.id}
              command={cmd}
              onDismiss={() => dismissCommand(cmd.id)}
              canDismiss={false}
            />
          ))}
        </div>
      )}

      {/* Downloads in progress */}
      {downloadItems.length > 0 && (
        <div className="space-y-2">
          {downloadItems.map((item) => (
            <QueueItemCard
              key={item.id}
              item={item}
              onRemove={() => handleRemove(item)}
              onRetry={item.status === "failed" ? () => retryMutation.mutate(item.id) : undefined}
              isRemoving={removeMutation.isPending}
              isRetrying={retryMutation.isPending}
            />
          ))}
        </div>
      )}

      {/* Finished searches (completed/failed) */}
      {finishedCommands.length > 0 && (
        <div className="space-y-2">
          {finishedCommands.map((cmd) => (
            <SearchCommandCard
              key={cmd.id}
              command={cmd}
              onDismiss={() => dismissCommand(cmd.id)}
              canDismiss={true}
            />
          ))}
        </div>
      )}

      {/* Delete confirmation modal */}
      {itemToRemove && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4 border">
            <h3 className="text-lg font-semibold mb-2">Remove from Queue?</h3>
            <p className="text-muted-foreground mb-4">
              Are you sure you want to remove "{itemToRemove.bookTitle}" from the download queue?
            </p>
            <div className="flex gap-2 justify-end">
              <Button variant="outline" onClick={() => setItemToRemove(null)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={confirmRemove}
                disabled={removeMutation.isPending}
              >
                {removeMutation.isPending ? "Removing..." : "Remove"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
