import { useState } from "react"
import { useDownloadQueue, useRemoveFromQueue } from "../../hooks/useWanted"
import { Button, Card, CardContent, Progress } from "../ui"
import type { QueueItem } from "../../lib/api"

export function QueueList() {
  const { data: queue, isLoading, error, refetch } = useDownloadQueue()
  const removeMutation = useRemoveFromQueue()
  const [itemToRemove, setItemToRemove] = useState<QueueItem | null>(null)

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

  const items = queue || []

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <p className="text-muted-foreground">
            {items.length} {items.length === 1 ? "item" : "items"} in queue
          </p>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isLoading}>
          Refresh
        </Button>
      </div>

      {/* Empty state */}
      {items.length === 0 && (
        <Card>
          <CardContent className="p-12 text-center">
            <div className="text-4xl mb-4">📭</div>
            <h3 className="text-lg font-semibold mb-2">Queue is empty</h3>
            <p className="text-muted-foreground">
              No downloads in progress. Search for books to start downloading.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Queue items */}
      {items.length > 0 && (
        <div className="space-y-3">
          {items.map((item) => (
            <QueueItemCard
              key={item.id}
              item={item}
              onRemove={() => handleRemove(item)}
              isRemoving={removeMutation.isPending}
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

interface QueueItemCardProps {
  item: QueueItem
  onRemove: () => void
  isRemoving: boolean
}

function QueueItemCard({ item, onRemove, isRemoving }: QueueItemCardProps) {
  const statusColors: Record<string, string> = {
    queued: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200",
    downloading: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
    paused: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
    completed: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
    seeding: "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200",
    failed: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
    extracted: "bg-teal-100 text-teal-800 dark:bg-teal-900 dark:text-teal-200",
    processing: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200",
  }

  const formatSize = (bytes: number) => {
    if (bytes === 0) return "0 B"
    const k = 1024
    const sizes = ["B", "KB", "MB", "GB"]
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
  }

  const formatSpeed = (bytesPerSecond: number) => {
    if (bytesPerSecond === 0) return ""
    return `${formatSize(bytesPerSecond)}/s`
  }

  const formatETA = (seconds: number) => {
    if (seconds <= 0) return ""
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    if (hours > 0) return `${hours}h ${minutes}m`
    if (minutes > 0) return `${minutes}m ${secs}s`
    return `${secs}s`
  }

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-medium truncate">{item.bookTitle}</h3>
              <span className={`text-xs px-2 py-0.5 rounded ${statusColors[item.status] || statusColors.queued}`}>
                {item.status}
              </span>
              <span className="text-xs text-muted-foreground uppercase">{item.format}</span>
            </div>
            <p className="text-xs text-muted-foreground truncate">{item.title}</p>

            {/* Progress bar for active downloads */}
            {item.status === "downloading" && (
              <div className="mt-2">
                <div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
                  <span>{Math.round(item.progress)}%</span>
                  <span>{item.size > 0 && formatSize(item.size)}</span>
                </div>
                <Progress value={item.progress} className="h-2" />
              </div>
            )}

            {/* Completed info */}
            {item.status === "completed" && (
              <p className="text-xs text-muted-foreground mt-1">
                {formatSize(item.size)} • Downloaded
              </p>
            )}

            {/* Error message */}
            {item.status === "failed" && (
              <p className="text-xs text-destructive mt-1">Download failed - file may require authorization</p>
            )}

            {/* Client info */}
            <p className="text-xs text-muted-foreground mt-1">{item.clientName}</p>
          </div>

          <Button
            variant="ghost"
            size="sm"
            onClick={onRemove}
            disabled={isRemoving}
            className="text-destructive hover:text-destructive"
          >
            ✕
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
