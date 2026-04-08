import { useState } from "react"
import { Link } from "@tanstack/react-router"
import { useBlocklist, useRemoveFromBlocklist } from "../../hooks/useWanted"
import { Button, Card, CardContent } from "../ui"
import type { BlocklistItem } from "../../lib/api"

export function BlocklistList() {
  const { data: blocklist, isLoading, error, refetch } = useBlocklist()
  const removeMutation = useRemoveFromBlocklist()
  const [itemToRemove, setItemToRemove] = useState<BlocklistItem | null>(null)

  const handleRemove = (item: BlocklistItem) => {
    setItemToRemove(item)
  }

  const confirmRemove = async () => {
    if (!itemToRemove) return
    try {
      await removeMutation.mutateAsync(itemToRemove.id)
      setItemToRemove(null)
    } catch (err) {
      console.error("Failed to remove from blocklist:", err)
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
          <p className="text-destructive">Failed to load blocklist</p>
          <Button variant="outline" onClick={() => refetch()} className="mt-4">
            Try Again
          </Button>
        </CardContent>
      </Card>
    )
  }

  const items = blocklist || []

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleString()
  }

  const formatRelativeTime = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffMins < 1) return "just now"
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return date.toLocaleDateString()
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <p className="text-muted-foreground">
          {items.length} {items.length === 1 ? "release" : "releases"} blocked
        </p>
        <Button variant="outline" onClick={() => refetch()} disabled={isLoading}>
          Refresh
        </Button>
      </div>

      {/* Info message */}
      <Card>
        <CardContent className="p-4 text-sm text-muted-foreground">
          <p>
            Blocked releases will not be downloaded again when searching for books. 
            Remove items from the blocklist to allow them to be downloaded in future searches.
          </p>
        </CardContent>
      </Card>

      {/* Empty state */}
      {items.length === 0 && (
        <Card>
          <CardContent className="p-12 text-center">
            <div className="text-4xl mb-4">🚫</div>
            <h3 className="text-lg font-semibold mb-2">Blocklist is empty</h3>
            <p className="text-muted-foreground">
              No releases have been blocked. Failed downloads can be added to the blocklist to prevent retry.
            </p>
          </CardContent>
        </Card>
      )}

      {/* Blocklist items */}
      {items.length > 0 && (
        <div className="space-y-2">
          {items.map((item) => (
            <Card key={item.id}>
              <CardContent className="p-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <Link
                        to="/book/$bookId"
                        params={{ bookId: String(item.bookId) }}
                        className="font-medium hover:underline truncate"
                      >
                        {item.bookTitle || "Unknown Book"}
                      </Link>
                    </div>
                    {item.authorName && (
                      <p className="text-xs text-muted-foreground">by {item.authorName}</p>
                    )}
                    <p className="text-xs text-muted-foreground truncate mt-1">
                      {item.sourceTitle}
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      {item.quality && (
                        <span className="text-xs bg-muted px-2 py-0.5 rounded uppercase">
                          {item.quality}
                        </span>
                      )}
                      {item.reason && (
                        <span className="text-xs text-destructive">
                          {item.reason}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground" title={formatDate(item.date)}>
                      {formatRelativeTime(item.date)}
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleRemove(item)}
                      disabled={removeMutation.isPending}
                      title="Remove from blocklist (allow future downloads)"
                    >
                      ✕
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Remove confirmation modal */}
      {itemToRemove && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4 border">
            <h3 className="text-lg font-semibold mb-2">Remove from Blocklist?</h3>
            <p className="text-muted-foreground mb-4">
              This will allow "{itemToRemove.sourceTitle}" to be downloaded again in future searches.
            </p>
            <div className="flex gap-2 justify-end">
              <Button variant="outline" onClick={() => setItemToRemove(null)}>
                Cancel
              </Button>
              <Button
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
