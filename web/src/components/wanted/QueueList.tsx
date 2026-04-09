import { useState, useEffect } from "react"
import { Link } from "@tanstack/react-router"
import { useDownloadQueue, useRemoveFromQueue, useActiveCommands, useRecentCommands } from "../../hooks/useWanted"
import { useRootFolders } from "../../hooks/useRootFolders"
import { Button, Card, CardContent, Progress } from "../ui"
import type { QueueItem, ActiveCommand } from "../../lib/api"

const DISMISSED_STORAGE_KEY = "bookaneer-dismissed-commands"

// Get dismissed command IDs from localStorage
function getDismissedCommands(): Set<string> {
  try {
    const stored = localStorage.getItem(DISMISSED_STORAGE_KEY)
    if (stored) {
      return new Set(JSON.parse(stored))
    }
  } catch {
    // Ignore errors
  }
  return new Set()
}

// Save dismissed command IDs to localStorage
function saveDismissedCommands(dismissed: Set<string>) {
  try {
    localStorage.setItem(DISMISSED_STORAGE_KEY, JSON.stringify([...dismissed]))
  } catch {
    // Ignore errors
  }
}

// Map command names to user-friendly descriptions
function getCommandDescription(command: ActiveCommand): { title: string; subtitle?: string; bookTitle?: string } {
  const name = command.name
  const payload = command.payload || {}
  const bookTitle = payload.bookTitle as string | undefined
  const authorName = payload.authorName as string | undefined

  switch (name) {
    case "AutomaticSearch":
      return {
        bookTitle,
        title: bookTitle || "Unknown Book",
        subtitle: authorName ? `by ${authorName}` : undefined,
      }
    case "BookSearch":
      return {
        bookTitle,
        title: bookTitle || "Unknown Book",
        subtitle: authorName ? `by ${authorName}` : undefined,
      }
    case "MissingBookSearch":
      return {
        title: "All Missing Books",
        subtitle: "Searching all monitored books",
      }
    case "DownloadGrab":
      return {
        bookTitle,
        title: bookTitle || "Unknown Book",
        subtitle: authorName ? `by ${authorName}` : undefined,
      }
    default:
      return { title: name, subtitle: undefined }
  }
}

function getStatusInfo(status: string, hasError: boolean) {
  if (status === "running" || status === "queued") {
    return { label: "Searching", color: "bg-blue-500", icon: "🔍", spinning: true }
  }
  if (status === "failed" || hasError) {
    return { label: "Not Found", color: "bg-red-500", icon: "✕", spinning: false }
  }
  return { label: "Found", color: "bg-green-500", icon: "✓", spinning: false }
}

export function QueueList() {
  const { data: queue, isLoading, error, refetch } = useDownloadQueue()
  const { data: activeCommands } = useActiveCommands()
  const { data: recentCommands } = useRecentCommands(50)
  const { data: rootFolders } = useRootFolders()
  const removeMutation = useRemoveFromQueue()
  const [itemToRemove, setItemToRemove] = useState<QueueItem | null>(null)
  const [dismissedCommands, setDismissedCommands] = useState<Set<string>>(getDismissedCommands)

  const hasRootFolder = rootFolders && rootFolders.length > 0

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
    setDismissedCommands(prev => new Set([...prev, commandId]))
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
  const bookCommands = (recentCommands || []).filter(cmd => 
    ["AutomaticSearch", "BookSearch", "MissingBookSearch", "DownloadGrab"].includes(cmd.name) &&
    !dismissedCommands.has(cmd.id)
  )

  // Active commands (still running)
  const activeBookCommands = (activeCommands || []).filter(cmd =>
    ["AutomaticSearch", "BookSearch", "MissingBookSearch", "DownloadGrab"].includes(cmd.name)
  )

  // Completed/failed commands
  const finishedCommands = bookCommands.filter(cmd => 
    cmd.status === "completed" || cmd.status === "failed"
  )

  const failedDownloads = downloadItems.filter(item => item.status === "failed")

  const handleClearAllFailed = async () => {
    for (const item of failedDownloads) {
      try {
        await removeMutation.mutateAsync(item.id)
      } catch (err) {
        console.error("Failed to remove item:", err)
      }
    }
  }

  const clearAllFinishedCommands = () => {
    finishedCommands.forEach(cmd => {
      dismissCommand(cmd.id)
    })
  }

  const hasActiveItems = activeBookCommands.length > 0 || downloadItems.length > 0
  const hasFinishedItems = finishedCommands.length > 0
  const isEmpty = !hasActiveItems && !hasFinishedItems

  return (
    <div className="space-y-6">
      {/* Warning: No root folder configured */}
      {!hasRootFolder && (
        <Card className="border-yellow-500/50 bg-yellow-500/10">
          <CardContent className="p-4">
            <div className="flex items-start gap-3">
              <span className="text-xl">⚠️</span>
              <div>
                <h4 className="font-medium text-yellow-600 dark:text-yellow-400">No Root Folder Configured</h4>
                <p className="text-sm text-muted-foreground mt-1">
                  Downloads will fail because there's no folder configured to save books.
                </p>
                <Link to="/settings" className="text-sm text-primary hover:underline mt-2 inline-block">
                  Go to Settings to add a root folder →
                </Link>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <p className="text-muted-foreground">
            {activeBookCommands.length > 0 && (
              <span>{activeBookCommands.length} searching</span>
            )}
            {downloadItems.length > 0 && (
              <span className="ml-2">{downloadItems.length} downloading</span>
            )}
            {hasFinishedItems && (
              <span className="ml-2">{finishedCommands.length} {finishedCommands.length === 1 ? "result" : "results"}</span>
            )}
            {isEmpty && "No activity"}
          </p>
        </div>
        <div className="flex items-center gap-2">
          {finishedCommands.length > 1 && (
            <Button 
              variant="outline" 
              size="sm"
              onClick={clearAllFinishedCommands}
            >
              Clear All Results
            </Button>
          )}
          {failedDownloads.length > 1 && (
            <Button 
              variant="outline" 
              size="sm"
              onClick={handleClearAllFailed}
              disabled={removeMutation.isPending}
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
            <div className="text-4xl mb-4">📭</div>
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
              isRemoving={removeMutation.isPending}
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

// Search command card component
interface SearchCommandCardProps {
  command: ActiveCommand
  onDismiss: () => void
  canDismiss: boolean
}

function SearchCommandCard({ command, onDismiss, canDismiss }: SearchCommandCardProps) {
  const { title, subtitle } = getCommandDescription(command)
  const hasError = command.status === "failed" || 
    (command.result?.error != null) ||
    (typeof command.result?.message === "string" && command.result.message.includes("no suitable"))
  
  const statusInfo = getStatusInfo(command.status, hasError)
  
  const errorMsg = command.result?.error 
    ? String(command.result.error) 
    : command.result?.message
      ? String(command.result.message)
      : null

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
  }

  const bookId = command.payload?.bookId as number | undefined

  return (
    <Card className={hasError && command.status !== "running" ? "border-red-500/30" : ""}>
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 flex-1 min-w-0">
            {/* Status indicator */}
            <div className="flex-shrink-0 mt-0.5">
              {statusInfo.spinning ? (
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-primary" />
              ) : (
                <span className={`flex items-center justify-center h-5 w-5 rounded-full text-white text-xs ${statusInfo.color}`}>
                  {statusInfo.icon}
                </span>
              )}
            </div>
            
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h3 className="font-medium truncate">{title}</h3>
                <span className={`text-xs px-2 py-0.5 rounded ${
                  statusInfo.spinning 
                    ? "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                    : hasError 
                      ? "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
                      : "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                }`}>
                  {statusInfo.label}
                </span>
              </div>
              {subtitle && (
                <p className="text-sm text-muted-foreground">{subtitle}</p>
              )}
              {errorMsg && !statusInfo.spinning && (
                <p className="text-sm text-red-400 mt-1" title={errorMsg}>
                  {errorMsg}
                </p>
              )}
              <p className="text-xs text-muted-foreground mt-1">
                {formatTime(command.endedAt || command.startedAt || command.queuedAt)}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-1 flex-shrink-0">
            {/* Link to book for retry */}
            {bookId && hasError && !statusInfo.spinning && (
              <Link to="/book/$bookId" params={{ bookId: String(bookId) }}>
                <Button variant="outline" size="sm" title="Open book page to search alternatives">
                  Search Again
                </Button>
              </Link>
            )}
            {canDismiss && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onDismiss}
                className="text-muted-foreground hover:text-foreground"
                title="Remove from list"
              >
                ✕
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
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

  return (
    <Card className={item.status === "failed" ? "border-red-500/30" : ""}>
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 flex-1 min-w-0">
            {/* Status indicator */}
            <div className="flex-shrink-0 mt-0.5">
              {item.status === "downloading" ? (
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-primary" />
              ) : item.status === "completed" ? (
                <span className="flex items-center justify-center h-5 w-5 rounded-full bg-green-500 text-white text-xs">✓</span>
              ) : item.status === "failed" ? (
                <span className="flex items-center justify-center h-5 w-5 rounded-full bg-red-500 text-white text-xs">✕</span>
              ) : (
                <span className="flex items-center justify-center h-5 w-5 rounded-full bg-gray-500 text-white text-xs">⏳</span>
              )}
            </div>

            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <h3 className="font-medium truncate">{item.bookTitle}</h3>
                <span className={`text-xs px-2 py-0.5 rounded ${statusColors[item.status] || statusColors.queued}`}>
                  {item.status === "downloading" ? "Downloading" : item.status}
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
                <div className="mt-1">
                  <p className="text-xs text-destructive">Download failed - this file requires login or is unavailable</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Go to the book page to search for alternative download sources
                  </p>
                </div>
              )}

              {/* Client info */}
              <p className="text-xs text-muted-foreground mt-1">{item.clientName}</p>
            </div>
          </div>

          <div className="flex items-center gap-1 flex-shrink-0">
            {item.status === "failed" && (
              <Link to="/book/$bookId" params={{ bookId: String(item.bookId) }}>
                <Button variant="outline" size="sm" title="Open book page to search alternatives">
                  Search Again
                </Button>
              </Link>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={onRemove}
              disabled={isRemoving}
              className="text-muted-foreground hover:text-destructive"
              title="Remove from queue"
            >
              ✕
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
