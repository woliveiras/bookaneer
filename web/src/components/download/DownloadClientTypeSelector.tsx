import type { DownloadClientType } from "../../lib/api"
import { Button, Card, CardHeader, CardTitle, CardContent } from "../ui"
import { DOWNLOAD_CLIENT_PRESETS, type DownloadClientPreset } from "./download-client-presets"

interface DownloadClientTypeSelectorProps {
  onSelect: (type: DownloadClientType, preset?: DownloadClientPreset) => void
  onCancel: () => void
}

export function DownloadClientTypeSelector({ onSelect, onCancel }: DownloadClientTypeSelectorProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Add Download Client</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <p className="text-sm text-muted-foreground">
          Select a download client to integrate with Bookaneer.
        </p>

        {/* Usenet Section */}
        <div className="space-y-3">
          <h4 className="font-medium text-sm text-muted-foreground uppercase tracking-wide">Usenet</h4>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
            {DOWNLOAD_CLIENT_PRESETS.usenet.map((preset) => (
              <button
                key={preset.id}
                type="button"
                onClick={() => onSelect(preset.type, preset)}
                className="p-4 rounded-lg border border-border bg-card hover:bg-accent hover:border-primary transition-colors text-left group"
              >
                <div className="font-medium group-hover:text-primary">{preset.name}</div>
                <div className="text-xs text-muted-foreground mt-1">{preset.description}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Torrents Section */}
        <div className="space-y-3">
          <h4 className="font-medium text-sm text-muted-foreground uppercase tracking-wide">Torrents</h4>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
            {DOWNLOAD_CLIENT_PRESETS.torrents.map((preset) => (
              <button
                key={preset.id}
                type="button"
                onClick={() => onSelect(preset.type, preset)}
                className="p-4 rounded-lg border border-border bg-card hover:bg-accent hover:border-primary transition-colors text-left group"
              >
                <div className="font-medium group-hover:text-primary">{preset.name}</div>
                <div className="text-xs text-muted-foreground mt-1">{preset.description}</div>
              </button>
            ))}
          </div>
        </div>

        <div className="flex justify-end pt-2">
          <Button variant="outline" onClick={onCancel}>
            Close
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
