import type { ActiveCommand } from "../../lib/api"

const DISMISSED_STORAGE_KEY = "bookaneer-dismissed-commands"

// Get dismissed command IDs from localStorage
export function getDismissedCommands(): Set<string> {
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
export function saveDismissedCommands(dismissed: Set<string>) {
  try {
    localStorage.setItem(DISMISSED_STORAGE_KEY, JSON.stringify([...dismissed]))
  } catch {
    // Ignore errors
  }
}

// Map command names to user-friendly descriptions
export function getCommandDescription(command: ActiveCommand): { title: string; subtitle?: string; bookTitle?: string } {
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

export function getStatusInfo(status: string, hasError: boolean) {
  if (status === "running" || status === "queued") {
    return { label: "Searching...", color: "bg-blue-500", icon: "search", spinning: true }
  }
  if (status === "failed" || hasError) {
    return { label: "Not Found", color: "bg-red-500", icon: "✕", spinning: false }
  }
  return { label: "Found", color: "bg-green-500", icon: "✓", spinning: false }
}
