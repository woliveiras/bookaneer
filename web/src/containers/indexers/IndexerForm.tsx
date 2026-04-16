import { useState } from "react"
import type { IndexerPreset, IndexerType } from "../../components/indexers/indexer-presets"
import { Button, Card, CardContent, CardHeader, CardTitle, Input, Label } from "../../components/ui"
import { useTestIndexer } from "../../hooks/useIndexers"
import type { CreateIndexerInput, Indexer } from "../../lib/api"

interface IndexerFormProps {
  indexer?: Indexer
  indexerType: IndexerType
  preset?: IndexerPreset
  onSubmit: (data: CreateIndexerInput) => void
  onCancel: () => void
  isLoading?: boolean
}

export function IndexerForm({
  indexer,
  indexerType,
  preset,
  onSubmit,
  onCancel,
  isLoading,
}: IndexerFormProps) {
  const [formData, setFormData] = useState<CreateIndexerInput>({
    name: indexer?.name ?? preset?.name ?? (indexerType === "newznab" ? "Newznab" : "Torznab"),
    type: indexer?.type ?? indexerType,
    baseUrl: indexer?.baseUrl ?? preset?.baseUrl ?? "",
    apiPath: indexer?.apiPath ?? preset?.apiPath ?? "/api",
    apiKey: indexer?.apiKey ?? "",
    categories: indexer?.categories ?? preset?.categories ?? "",
    priority: indexer?.priority ?? 25,
    enabled: indexer?.enabled ?? true,
    enableRss: indexer?.enableRss ?? true,
    enableAutomaticSearch: indexer?.enableAutomaticSearch ?? true,
    enableInteractiveSearch: indexer?.enableInteractiveSearch ?? true,
    additionalParameters: indexer?.additionalParameters ?? "",
    minimumSeeders: indexer?.minimumSeeders ?? 1,
    seedRatio: indexer?.seedRatio ?? null,
    seedTime: indexer?.seedTime ?? null,
  })
  const testIndexer = useTestIndexer()
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)

  const handleTest = async () => {
    setTestResult(null)
    try {
      const result = await testIndexer.mutateAsync(formData)
      setTestResult(result)
    } catch {
      setTestResult({ success: false, message: "Connection failed" })
    }
  }

  const isTorznab = formData.type === "torznab"
  const title = indexer
    ? `Edit Indexer - ${indexer.name}`
    : `Add Indexer - ${indexerType === "newznab" ? "Newznab" : "Torznab"}`

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            onSubmit(formData)
          }}
          className="space-y-4"
        >
          {/* Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="My Indexer"
              required
            />
          </div>

          {/* Enable flags */}
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <input
                id="enableRss"
                type="checkbox"
                checked={formData.enableRss}
                onChange={(e) => setFormData({ ...formData, enableRss: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300"
              />
              <div>
                <Label htmlFor="enableRss">Enable RSS</Label>
                <p className="text-xs text-muted-foreground">
                  Will be used when Bookaneer periodically looks for releases via RSS Sync
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <input
                id="enableAutomaticSearch"
                type="checkbox"
                checked={formData.enableAutomaticSearch}
                onChange={(e) =>
                  setFormData({ ...formData, enableAutomaticSearch: e.target.checked })
                }
                className="h-4 w-4 rounded border-gray-300"
              />
              <div>
                <Label htmlFor="enableAutomaticSearch">Enable Automatic Search</Label>
                <p className="text-xs text-muted-foreground">
                  Will be used when automatic searches are performed via the UI or by Bookaneer
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <input
                id="enableInteractiveSearch"
                type="checkbox"
                checked={formData.enableInteractiveSearch}
                onChange={(e) =>
                  setFormData({ ...formData, enableInteractiveSearch: e.target.checked })
                }
                className="h-4 w-4 rounded border-gray-300"
              />
              <div>
                <Label htmlFor="enableInteractiveSearch">Enable Interactive Search</Label>
                <p className="text-xs text-muted-foreground">
                  Will be used when interactive search is used
                </p>
              </div>
            </div>
          </div>

          {/* URL and API Path */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="sm:col-span-2 space-y-2">
              <Label htmlFor="baseUrl">URL</Label>
              <Input
                id="baseUrl"
                type="url"
                value={formData.baseUrl}
                onChange={(e) => setFormData({ ...formData, baseUrl: e.target.value })}
                placeholder="https://indexer.example.com"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="apiPath" className="text-amber-500">
                API Path
              </Label>
              <Input
                id="apiPath"
                value={formData.apiPath}
                onChange={(e) => setFormData({ ...formData, apiPath: e.target.value })}
                placeholder="/api"
              />
              <p className="text-xs text-muted-foreground">Path to the API, usually /api</p>
            </div>
          </div>

          {/* API Key */}
          <div className="space-y-2">
            <Label htmlFor="apiKey">API Key</Label>
            <Input
              id="apiKey"
              type="password"
              value={formData.apiKey}
              onChange={(e) => setFormData({ ...formData, apiKey: e.target.value })}
              placeholder="Your API key"
            />
          </div>

          {/* Categories */}
          <div className="space-y-2">
            <Label htmlFor="categories">Categories</Label>
            <Input
              id="categories"
              value={formData.categories}
              onChange={(e) => setFormData({ ...formData, categories: e.target.value })}
              placeholder="7000,7010,7020"
            />
            <p className="text-xs text-muted-foreground">
              Comma-separated category IDs (e.g., 7000 for Books)
            </p>
          </div>

          {/* Additional Parameters */}
          <div className="space-y-2">
            <Label htmlFor="additionalParameters" className="text-amber-500">
              Additional Parameters
            </Label>
            <Input
              id="additionalParameters"
              value={formData.additionalParameters}
              onChange={(e) => setFormData({ ...formData, additionalParameters: e.target.value })}
              placeholder=""
            />
            <p className="text-xs text-muted-foreground">Additional Newznab/Torznab parameters</p>
          </div>

          {/* Torznab-specific fields */}
          {isTorznab && (
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="minimumSeeders" className="text-amber-500">
                  Minimum Seeders
                </Label>
                <Input
                  id="minimumSeeders"
                  type="number"
                  min={0}
                  value={formData.minimumSeeders}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      minimumSeeders: parseInt(e.target.value, 10) || 0,
                    })
                  }
                />
                <p className="text-xs text-muted-foreground">Minimum number of seeders required</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="seedRatio">Seed Ratio</Label>
                <Input
                  id="seedRatio"
                  type="number"
                  step="0.1"
                  min={0}
                  value={formData.seedRatio ?? ""}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      seedRatio: e.target.value ? parseFloat(e.target.value) : null,
                    })
                  }
                  placeholder=""
                />
                <p className="text-xs text-muted-foreground">Empty uses download client default</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="seedTime">Seed Time (minutes)</Label>
                <Input
                  id="seedTime"
                  type="number"
                  min={0}
                  value={formData.seedTime ?? ""}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      seedTime: e.target.value ? parseInt(e.target.value, 10) : null,
                    })
                  }
                  placeholder=""
                />
                <p className="text-xs text-muted-foreground">Empty uses download client default</p>
              </div>
            </div>
          )}

          {/* Priority */}
          <div className="space-y-2">
            <Label htmlFor="priority" className="text-amber-500">
              Indexer Priority
            </Label>
            <Input
              id="priority"
              type="number"
              min={1}
              max={50}
              value={formData.priority}
              onChange={(e) =>
                setFormData({ ...formData, priority: parseInt(e.target.value, 10) || 25 })
              }
            />
            <p className="text-xs text-muted-foreground">
              Indexer Priority from 1 (Highest) to 50 (Lowest). Default: 25. Used as tiebreaker for
              otherwise equal releases.
            </p>
          </div>

          {/* Enabled */}
          <div className="flex items-center gap-2">
            <input
              id="enabled"
              type="checkbox"
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="enabled">Enabled</Label>
          </div>

          {/* Test result */}
          {testResult && (
            <div
              className={`p-3 rounded-md text-sm ${
                testResult.success
                  ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
                  : "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400"
              }`}
            >
              {testResult.message}
            </div>
          )}

          {/* Actions */}
          <div className="flex flex-wrap justify-end gap-2 pt-4 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={handleTest}
              disabled={testIndexer.isPending || !formData.baseUrl}
            >
              {testIndexer.isPending ? "Testing..." : "Test"}
            </Button>
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Saving..." : "Save"}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
