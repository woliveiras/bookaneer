import { useState } from "react"
import type { GrabMeta } from "../../hooks/useBookRelease"
import type { DigitalLibraryResult, SearchResult } from "../../lib/api"
import { formatBytes } from "../../lib/format"
import type { ColumnConfig } from "../../lib/types"
import { Badge, Button, Card, CardContent } from "../ui"
import { DynamicCell } from "./DynamicColumns"

// ─── DownloadResult ──────────────────────────────────────────────────────────

interface DownloadResultProps {
  result: SearchResult
  onGrab: (url: string, title: string, size: number, meta?: GrabMeta) => Promise<void>
  isGrabbing: boolean
  columnConfig?: ColumnConfig
  fromExpanded?: boolean
}

export function DownloadResult({
  result,
  onGrab,
  isGrabbing,
  columnConfig,
  fromExpanded,
}: DownloadResultProps) {
  const [grabbing, setGrabbing] = useState(false)
  const [grabbed, setGrabbed] = useState(false)

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

  return (
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
          <Button
            size="sm"
            onClick={handleGrab}
            disabled={grabbing || isGrabbing || grabbed}
            variant={grabbed ? "secondary" : "default"}
          >
            {grabbing ? "Grabbing..." : grabbed ? "Grabbed" : "Grab"}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

interface LibraryResultProps {
  result: DigitalLibraryResult
  onGrab: (url: string, title: string, size: number, meta?: GrabMeta) => Promise<void>
  isGrabbing: boolean
  columnConfig?: ColumnConfig
  fromExpanded?: boolean
}

export function LibraryResult({
  result,
  onGrab,
  isGrabbing,
  columnConfig,
  fromExpanded,
}: LibraryResultProps) {
  const [grabbing, setGrabbing] = useState(false)
  const [grabbed, setGrabbed] = useState(false)

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
            <Button
              size="sm"
              onClick={handleGrab}
              disabled={grabbing || isGrabbing || grabbed}
              variant={grabbed ? "secondary" : "default"}
            >
              {grabbing ? "Grabbing..." : grabbed ? "Grabbed" : "Grab"}
            </Button>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
