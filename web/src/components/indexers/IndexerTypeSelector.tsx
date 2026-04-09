import { Button, Card, CardHeader, CardTitle, CardContent } from "../ui"
import { INDEXER_PRESETS, type IndexerType, type IndexerPreset } from "./indexer-presets"

interface IndexerTypeSelectorProps {
  onSelect: (type: IndexerType, preset?: IndexerPreset) => void
  onCancel: () => void
}

export function IndexerTypeSelector({ onSelect, onCancel }: IndexerTypeSelectorProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Add Indexer</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <p className="text-sm text-muted-foreground">
          Bookaneer supports any indexer that uses the Newznab/Torznab standard.
          Select a preset or choose Custom to configure manually.
        </p>

        {/* Usenet Section */}
        <div className="space-y-3">
          <h4 className="font-medium text-sm text-muted-foreground uppercase tracking-wide">Usenet</h4>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {INDEXER_PRESETS.usenet.map((preset) => (
              <button
                key={preset.id}
                type="button"
                onClick={() => onSelect(preset.type, preset)}
                className="p-4 rounded-lg border border-border bg-card hover:bg-accent hover:border-primary transition-colors text-left group"
              >
                <div className="font-medium group-hover:text-primary">{preset.name}</div>
                <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{preset.description}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Torrents Section */}
        <div className="space-y-3">
          <h4 className="font-medium text-sm text-muted-foreground uppercase tracking-wide">Torrents</h4>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
            {INDEXER_PRESETS.torrents.map((preset) => (
              <button
                key={preset.id}
                type="button"
                onClick={() => onSelect(preset.type, preset)}
                className="p-4 rounded-lg border border-border bg-card hover:bg-accent hover:border-primary transition-colors text-left group"
              >
                <div className="font-medium group-hover:text-primary">{preset.name}</div>
                <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{preset.description}</div>
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
