import { Link } from "@tanstack/react-router"
import type { ActiveCommand, QueueItem } from "../../lib/api"
import { Button, Card, CardContent, Progress } from "../ui"
import { DownloadingIcon, QueuedIcon, SearchingIcon } from "./QueueIcons"
import { getCommandDescription, getStatusInfo } from "./queueHelpers"

interface SearchCommandCardProps {
  command: ActiveCommand
  onDismiss: () => void
  canDismiss: boolean
}

export function SearchCommandCard({ command, onDismiss, canDismiss }: SearchCommandCardProps) {
  const { title, subtitle } = getCommandDescription(command)
  const hasError =
    command.status === "failed" ||
    command.result?.error != null ||
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
    <Card
      className={`relative overflow-hidden ${hasError && command.status !== "running" ? "border-red-500/30" : ""}`}
    >
      {/* Shimmer gradient overlay when searching */}
      {statusInfo.spinning && (
        <div
          className="absolute inset-0 -translate-x-full animate-[shimmer_2s_infinite]"
          style={{
            background:
              "linear-gradient(90deg, transparent 0%, rgba(59, 130, 246, 0.1) 50%, transparent 100%)",
          }}
        />
      )}
      <CardContent className="p-4 relative z-10">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 flex-1 min-w-0">
            {/* Status indicator */}
            <div className="shrink-0 mt-0.5">
              {statusInfo.spinning ? (
                <SearchingIcon />
              ) : (
                <span
                  className={`flex items-center justify-center h-6 w-6 rounded-full text-white text-xs ${statusInfo.color}`}
                >
                  {statusInfo.icon}
                </span>
              )}
            </div>

            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h3 className="font-medium truncate">{title}</h3>
                <span
                  className={`text-xs px-2 py-0.5 rounded ${
                    statusInfo.spinning
                      ? "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                      : hasError
                        ? "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
                        : "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                  }`}
                >
                  {statusInfo.label}
                </span>
              </div>
              {subtitle && <p className="text-sm text-muted-foreground">{subtitle}</p>}
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

          <div className="flex items-center gap-1 shrink-0">
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

export function QueueItemCard({ item, onRemove, isRemoving }: QueueItemCardProps) {
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
    return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`
  }

  const isActive = item.status === "queued" || item.status === "downloading"

  return (
    <Card
      className={`relative overflow-hidden ${item.status === "failed" ? "border-red-500/30" : ""}`}
    >
      {/* Shimmer gradient overlay when downloading or queued */}
      {isActive && (
        <div
          className="absolute inset-0 -translate-x-full animate-[shimmer_2s_infinite]"
          style={{
            background:
              item.status === "downloading"
                ? "linear-gradient(90deg, transparent 0%, rgba(59, 130, 246, 0.15) 50%, transparent 100%)"
                : "linear-gradient(90deg, transparent 0%, rgba(245, 158, 11, 0.1) 50%, transparent 100%)",
          }}
        />
      )}
      <CardContent className="p-4 relative z-10">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 flex-1 min-w-0">
            {/* Status indicator */}
            <div className="shrink-0 mt-0.5">
              {item.status === "downloading" ? (
                <DownloadingIcon />
              ) : item.status === "completed" ? (
                <span className="flex items-center justify-center h-6 w-6 rounded-full bg-green-500 text-white text-xs">
                  ✓
                </span>
              ) : item.status === "failed" ? (
                <span className="flex items-center justify-center h-6 w-6 rounded-full bg-red-500 text-white text-xs">
                  ✕
                </span>
              ) : (
                <QueuedIcon />
              )}
            </div>

            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <h3 className="font-medium truncate">{item.bookTitle}</h3>
                <span
                  className={`text-xs px-2 py-0.5 rounded ${statusColors[item.status] || statusColors.queued}`}
                >
                  {item.status === "downloading"
                    ? "Downloading..."
                    : item.status === "queued"
                      ? "Queued"
                      : item.status === "completed"
                        ? "Completed"
                        : item.status === "failed"
                          ? "Failed"
                          : item.status}
                </span>
                <span className="text-xs text-muted-foreground uppercase">{item.format}</span>
              </div>
              <p className="text-xs text-muted-foreground truncate">{item.title}</p>

              {/* Progress bar for queued items - indeterminate */}
              {item.status === "queued" && (
                <div className="mt-2">
                  <div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
                    <span>Waiting to download...</span>
                  </div>
                  <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-amber-500 rounded-full animate-pulse"
                      style={{ width: "30%" }}
                    />
                  </div>
                </div>
              )}

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
                  <p className="text-xs text-destructive">
                    Download failed - this file requires login or is unavailable
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Go to the book page to search for alternative download sources
                  </p>
                </div>
              )}

              {/* Client info */}
              <p className="text-xs text-muted-foreground mt-1">{item.clientName}</p>
            </div>
          </div>

          <div className="flex items-center gap-1 shrink-0">
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
