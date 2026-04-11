import { Link } from "@tanstack/react-router"
import { AlertTriangle } from "lucide-react"
import { useRootFolders } from "../../hooks/useRootFolders"
import { Card, CardContent } from "../ui"

/**
 * Global warning banner that appears when no root folder is configured.
 * Shows on all pages except Settings to remind users to configure essential settings.
 */
export function RootFolderWarning() {
  const { data: rootFolders, isLoading } = useRootFolders()

  // Don't show while loading
  if (isLoading) return null

  // Don't show if root folder is configured
  const hasRootFolder = rootFolders && rootFolders.length > 0
  if (hasRootFolder) return null

  return (
    <Card className="border-yellow-500/50 bg-yellow-500/10 mb-6">
      <CardContent className="p-4">
        <div className="flex items-start gap-3">
          <AlertTriangle className="w-5 h-5 text-yellow-600" />
          <div>
            <h4 className="font-medium text-yellow-600 dark:text-yellow-400">
              No Root Folder Configured
            </h4>
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
  )
}
