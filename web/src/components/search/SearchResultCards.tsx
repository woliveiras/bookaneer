import { useState } from "react"
import { toast } from "sonner"
import type { GrabMeta } from "../../hooks/useBookRelease"
import type { DigitalLibraryResult, SearchResult } from "../../lib/api"
import { formatBytes } from "../../lib/format"
import type { ColumnConfig } from "../../lib/types"
import { Badge, Button, Card, CardContent } from "../ui"
import { DynamicCell } from "./DynamicColumns"

// ─── Shared Import Confirmation Dialog ──────────────────────────────────────

interface ImportConfirmDialogProps {
  title: string
  onCancel: () => void
  onConfirm: () => void
}

function ImportConfirmDialog({ title, onCancel, onConfirm }: ImportConfirmDialogProps) {
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4 border">
        <h3 className="text-lg font-semibold mb-2">Replace existing file?</h3>
        <p className="text-muted-foreground mb-4">
          The current file for <strong>"{title}"</strong> will be removed and replaced by this
          download. This action cannot be undone.
        </p>
        <div className="flex gap-2 justify-end">
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm}>
            Replace
          </Button>
        </div>
      </div>
    </div>
  )
}

// ─── DownloadResult ──────────────────────────────────────────────────────────

interface DownloadResultProps {
  result: SearchResult
  onGrab: (url: string, title: string, size: number, meta?: GrabMeta) => Promise<void>
  isGrabbing: boolean
  columnConfig?: ColumnConfig
  fromExpanded?: boolean
  hasExistingFile?: boolean
}

export function DownloadResult({
  result,
  onGrab,
  isGrabbing,
  columnConfig,
  fromExpanded,
  hasExistingFile = false,
}: DownloadResultProps) {
  const [grabbing, setGrabbing] = useState(false)
  const [grabbed, setGrabbed] = useState(false)
  const [imported, setImported] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)

  const handleGrab = async () => {
    setGrabbing(true)
    try {
      await onGrab(result.downloadUrl, result.title, result.size, {
        sourceType: "indexer",
        guid: result.guid,
        seeders: result.seeders,
        indexerId: result.indexerId,
        indexerName: result.indexerName,
      })
      setGrabbed(true)
    } finally {
      setGrabbing(false)
    }
  }

  const handleConfirmImport = () => {
    setShowConfirm(false)
    setImported(true)
    toast.success(`"${result.title}" will replace the existing file.`)
  }

  const needsImport = grabbed && hasExistingFile && !imported

  return (
    <>
      <Card>
        <CardContent className="py-3 px-4">
          <div className="flex justify-between items-center gap-4">
            <div className="flex-1 min-w-0">
              <h4 className="font-medium text-sm line-clamp-2">{result.title}</h4>
              {columnConfig ? (
                <div className="flex flex-wrap gap-1 mt-1">
                  {columnConfig.columns.map((col) => (
                    <DynamicCell
                      key={col.key}
                      column={col}
                      row={result as unknown as Record<string, unknown>}
                    />
                  ))}
                  {fromExpanded && (
                    <Badge variant="outline" className="text-xs border-violet-400 text-violet-500">
                      Expanded Search
                    </Badge>
                  )}
                </div>
              ) : (
                <div className="flex flex-wrap gap-1 mt-1">
                  <Badge variant="outline" className="text-xs">
                    {result.indexerName}
                  </Badge>
                  <Badge variant="secondary" className="text-xs">
                    {formatBytes(result.size)}
                  </Badge>
                  {result.seeders !== undefined && result.seeders > 0 && (
                    <Badge variant="default" className="bg-green-600 text-xs">
                      {result.seeders} seeds
                    </Badge>
                  )}
                  {fromExpanded && (
                    <Badge variant="outline" className="text-xs border-violet-400 text-violet-500">
                      Expanded Search
                    </Badge>
                  )}
                </div>
              )}
            </div>
            {needsImport ? (
              <Button size="sm" variant="outline" onClick={() => setShowConfirm(true)}>
                Import
              </Button>
            ) : (
              <Button
                size="sm"
                onClick={handleGrab}
                disabled={grabbing || isGrabbing || grabbed}
                variant={grabbed ? "secondary" : "default"}
              >
                {grabbing ? "Grabbing..." : grabbed ? "Grabbed" : "Grab"}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
      {showConfirm && (
        <ImportConfirmDialog
          title={result.title}
          onCancel={() => setShowConfirm(false)}
          onConfirm={handleConfirmImport}
        />
      )}
    </>
  )
}

// ─── LibraryResult ───────────────────────────────────────────────────────────

interface LibraryResultProps {
  result: DigitalLibraryResult
  onGrab: (url: string, title: string, size: number, meta?: GrabMeta) => Promise<void>
  isGrabbing: boolean
  columnConfig?: ColumnConfig
  fromExpanded?: boolean
  hasExistingFile?: boolean
}

export function LibraryResult({
  result,
  onGrab,
  isGrabbing,
  columnConfig,
  fromExpanded,
  hasExistingFile = false,
}: LibraryResultProps) {
  const [grabbing, setGrabbing] = useState(false)
  const [grabbed, setGrabbed] = useState(false)
  const [imported, setImported] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)

  const handleGrab = async () => {
    const url = result.downloadUrl || result.infoUrl
    if (!url) return

    setGrabbing(true)
    try {
      await onGrab(url, result.title, result.size)
      setGrabbed(true)
    } finally {
      setGrabbing(false)
    }
  }

  const handleConfirmImport = () => {
    setShowConfirm(false)
    setImported(true)
    toast.success(`"${result.title}" will replace the existing file.`)
  }

  const needsImport = grabbed && hasExistingFile && !imported

  return (
    <>
      <Card>
        <CardContent className="py-3 px-4">
        <div className="flex justify-between items-center gap-4">
          <div className="flex-1 min-w-0">
            <h4 className="font-medium text-sm line-clamp-2">{result.title}</h4>
            {result.authors && result.authors.length > 0 && (
              <p className="text-xs text-muted-foreground">{result.authors.join(", ")}</p>
            )}
            {columnConfig ? (
              <div className="flex flex-wrap gap-1 mt-1">
                {columnConfig.columns.map((col) => (
                  <DynamicCell
                    key={col.key}
                    column={col}
                    row={result as unknown as Record<string, unknown>}
                  />
                ))}
                {fromExpanded && (
                  <Badge variant="outline" className="text-xs border-violet-400 text-violet-500">
                    Expanded Search
                  </Badge>
                )}
              </div>
            ) : (
              <div className="flex flex-wrap gap-1 mt-1">
                <Badge variant="outline" className="text-xs">
                  {result.provider}
                </Badge>
                <Badge variant="secondary" className="text-xs uppercase">
                  {result.format}
                </Badge>
                {result.size > 0 && (
                  <Badge variant="secondary" className="text-xs">
                    {formatBytes(result.size)}
                  </Badge>
                )}
                {result.year && (
                  <Badge variant="secondary" className="text-xs">
                    {result.year}
                  </Badge>
                )}
                {result.score && (
                  <Badge variant="default" className="text-xs bg-primary">
                    Score: {result.score}
                  </Badge>
                )}
                {fromExpanded && (
                  <Badge variant="outline" className="text-xs border-violet-400 text-violet-500">
                    Expanded Search
                  </Badge>
                )}
              </div>
            )}
          </div>
          {needsImport ? (
            <Button size="sm" variant="outline" onClick={() => setShowConfirm(true)}>
              Import
            </Button>
          ) : (
            <Button
              size="sm"
              onClick={handleGrab}
              disabled={grabbing || isGrabbing || grabbed}
              variant={grabbed ? "secondary" : "default"}
            >
              {grabbing ? "Grabbing..." : grabbed ? "Grabbed" : "Grab"}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
    {showConfirm && (
      <ImportConfirmDialog
        title={result.title}
        onCancel={() => setShowConfirm(false)}
        onConfirm={handleConfirmImport}
      />
    )}
    </>
  )
}
