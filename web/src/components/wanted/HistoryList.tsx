import { useState } from "react"
import { Link } from "@tanstack/react-router"
import { useHistory } from "../../hooks/useWanted"
import { Button, Card, CardContent } from "../ui"
import type { HistoryEventType } from "../../lib/api"

const eventTypeLabels: Record<HistoryEventType, { label: string; color: string; icon: string }> = {
  grabbed: { label: "Grabbed", color: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200", icon: "⬇️" },
  downloadCompleted: { label: "Downloaded", color: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200", icon: "✅" },
  downloadFailed: { label: "Failed", color: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200", icon: "❌" },
  bookFileDeleted: { label: "Deleted", color: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200", icon: "🗑️" },
  bookFileRenamed: { label: "Renamed", color: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200", icon: "📝" },
  bookImported: { label: "Imported", color: "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200", icon: "📚" },
}

const eventTypeFilters: Array<{ value: HistoryEventType | ""; label: string }> = [
  { value: "", label: "All Events" },
  { value: "grabbed", label: "Grabbed" },
  { value: "downloadCompleted", label: "Downloaded" },
  { value: "downloadFailed", label: "Failed" },
  { value: "bookImported", label: "Imported" },
  { value: "bookFileDeleted", label: "Deleted" },
]

export function HistoryList() {
  const [eventTypeFilter, setEventTypeFilter] = useState<HistoryEventType | "">("")
  const { data: history, isLoading, error, refetch } = useHistory({ 
    limit: 100, 
    eventType: eventTypeFilter || undefined 
  })

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
          <p className="text-destructive">Failed to load history</p>
          <Button variant="outline" onClick={() => refetch()} className="mt-4">
            Try Again
          </Button>
        </CardContent>
      </Card>
    )
  }

  const items = history || []

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
      {/* Header with filter */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <p className="text-muted-foreground">
            {items.length} {items.length === 1 ? "event" : "events"}
          </p>
          <select
            value={eventTypeFilter}
            onChange={(e) => setEventTypeFilter(e.target.value as HistoryEventType | "")}
            className="bg-background border rounded px-3 py-1 text-sm"
          >
            {eventTypeFilters.map((filter) => (
              <option key={filter.value} value={filter.value}>
                {filter.label}
              </option>
            ))}
          </select>
        </div>
        <Button variant="outline" onClick={() => refetch()} disabled={isLoading}>
          Refresh
        </Button>
      </div>

      {/* Empty state */}
      {items.length === 0 && (
        <Card>
          <CardContent className="p-12 text-center">
            <div className="text-4xl mb-4">📜</div>
            <h3 className="text-lg font-semibold mb-2">No history</h3>
            <p className="text-muted-foreground">
              {eventTypeFilter ? "No events match this filter." : "No activity recorded yet."}
            </p>
          </CardContent>
        </Card>
      )}

      {/* History items */}
      {items.length > 0 && (
        <div className="space-y-2">
          {items.map((item) => {
            const eventInfo = eventTypeLabels[item.eventType] || { 
              label: item.eventType, 
              color: "bg-gray-100 text-gray-800", 
              icon: "📋" 
            }
            
            return (
              <Card key={item.id}>
                <CardContent className="p-4">
                  <div className="flex items-start gap-4">
                    <span className="text-2xl">{eventInfo.icon}</span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        {item.bookId ? (
                          <Link
                            to="/books/$bookId"
                            params={{ bookId: String(item.bookId) }}
                            className="font-medium hover:underline truncate"
                          >
                            {item.bookTitle || "Unknown Book"}
                          </Link>
                        ) : (
                          <span className="font-medium truncate">
                            {item.bookTitle || "Unknown Book"}
                          </span>
                        )}
                        <span className={`text-xs px-2 py-0.5 rounded ${eventInfo.color}`}>
                          {eventInfo.label}
                        </span>
                      </div>
                      {item.authorName && (
                        <p className="text-xs text-muted-foreground">by {item.authorName}</p>
                      )}
                      {item.sourceTitle && (
                        <p className="text-xs text-muted-foreground truncate mt-1">
                          {item.sourceTitle}
                        </p>
                      )}
                      {item.quality && (
                        <span className="text-xs text-muted-foreground uppercase">
                          {item.quality}
                        </span>
                      )}
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-muted-foreground" title={formatDate(item.date)}>
                        {formatRelativeTime(item.date)}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
