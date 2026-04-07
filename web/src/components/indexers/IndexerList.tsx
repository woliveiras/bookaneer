import { useState } from "react"
import { useIndexers, useDeleteIndexer, useTestIndexer, useCreateIndexer, useUpdateIndexer } from "../../hooks/useIndexers"
import type { Indexer, CreateIndexerInput } from "../../lib/api"
import { Button, Input, Label, Card, CardHeader, CardTitle, CardContent, Badge } from "../ui"

type IndexerType = "newznab" | "torznab"

// Preset configurations for popular ebook indexers
interface IndexerPreset {
  id: string
  name: string
  type: IndexerType
  baseUrl: string
  apiPath: string
  categories: string
  description: string
}

const INDEXER_PRESETS: { usenet: IndexerPreset[]; torrents: IndexerPreset[] } = {
  usenet: [
    {
      id: "nzbgeek",
      name: "NZBgeek",
      type: "newznab",
      baseUrl: "https://api.nzbgeek.info",
      apiPath: "/api",
      categories: "7000,7020,7030",
      description: "Popular Usenet indexer with ebooks",
    },
    {
      id: "drunkenslug",
      name: "DrunkenSlug",
      type: "newznab",
      baseUrl: "https://api.drunkenslug.com",
      apiPath: "/api",
      categories: "7000,7020,7030",
      description: "Usenet indexer with good ebook coverage",
    },
    {
      id: "nzbfinder",
      name: "NZBFinder",
      type: "newznab",
      baseUrl: "https://nzbfinder.ws",
      apiPath: "/api",
      categories: "7000,7020,7030",
      description: "Dutch Usenet indexer with ebooks",
    },
    {
      id: "newznab-custom",
      name: "Newznab",
      type: "newznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "7000",
      description: "Custom Newznab-compatible indexer",
    },
  ],
  torrents: [
    {
      id: "myanonamouse",
      name: "MyAnonamouse",
      type: "torznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "8000,8010",
      description: "Private tracker for ebooks (via Prowlarr/Jackett)",
    },
    {
      id: "bibliotik",
      name: "BiblioTik",
      type: "torznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "8000,8010",
      description: "Private ebook tracker (via Prowlarr/Jackett)",
    },
    {
      id: "prowlarr",
      name: "Prowlarr",
      type: "torznab",
      baseUrl: "http://localhost:9696",
      apiPath: "/1/api",
      categories: "",
      description: "Indexer manager/proxy for Servarr",
    },
    {
      id: "jackett",
      name: "Jackett",
      type: "torznab",
      baseUrl: "http://localhost:9117",
      apiPath: "/api/v2.0/indexers/all/results/torznab",
      categories: "",
      description: "Torznab proxy for many trackers",
    },
    {
      id: "torznab-custom",
      name: "Torznab",
      type: "torznab",
      baseUrl: "",
      apiPath: "/api",
      categories: "",
      description: "Custom Torznab-compatible indexer",
    },
  ],
}

interface IndexerTypeSelectorProps {
  onSelect: (type: IndexerType, preset?: IndexerPreset) => void
  onCancel: () => void
}

function IndexerTypeSelector({ onSelect, onCancel }: IndexerTypeSelectorProps) {
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

interface IndexerFormProps {
  indexer?: Indexer
  indexerType: IndexerType
  preset?: IndexerPreset
  onSubmit: (data: CreateIndexerInput) => void
  onCancel: () => void
  isLoading?: boolean
}

function IndexerForm({ indexer, indexerType, preset, onSubmit, onCancel, isLoading }: IndexerFormProps) {
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
  const title = indexer ? `Edit Indexer - ${indexer.name}` : `Add Indexer - ${indexerType === "newznab" ? "Newznab" : "Torznab"}`

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
                <p className="text-xs text-muted-foreground">Will be used when Bookaneer periodically looks for releases via RSS Sync</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <input
                id="enableAutomaticSearch"
                type="checkbox"
                checked={formData.enableAutomaticSearch}
                onChange={(e) => setFormData({ ...formData, enableAutomaticSearch: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300"
              />
              <div>
                <Label htmlFor="enableAutomaticSearch">Enable Automatic Search</Label>
                <p className="text-xs text-muted-foreground">Will be used when automatic searches are performed via the UI or by Bookaneer</p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <input
                id="enableInteractiveSearch"
                type="checkbox"
                checked={formData.enableInteractiveSearch}
                onChange={(e) => setFormData({ ...formData, enableInteractiveSearch: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300"
              />
              <div>
                <Label htmlFor="enableInteractiveSearch">Enable Interactive Search</Label>
                <p className="text-xs text-muted-foreground">Will be used when interactive search is used</p>
              </div>
            </div>
          </div>

          {/* URL and API Path */}
          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
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
              <Label htmlFor="apiPath" className="text-amber-500">API Path</Label>
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
            <p className="text-xs text-muted-foreground">Comma-separated category IDs (e.g., 7000 for Books)</p>
          </div>

          {/* Additional Parameters */}
          <div className="space-y-2">
            <Label htmlFor="additionalParameters" className="text-amber-500">Additional Parameters</Label>
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
            <>
              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="minimumSeeders" className="text-amber-500">Minimum Seeders</Label>
                  <Input
                    id="minimumSeeders"
                    type="number"
                    min={0}
                    value={formData.minimumSeeders}
                    onChange={(e) => setFormData({ ...formData, minimumSeeders: parseInt(e.target.value, 10) || 0 })}
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
                    onChange={(e) => setFormData({ ...formData, seedRatio: e.target.value ? parseFloat(e.target.value) : null })}
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
                    onChange={(e) => setFormData({ ...formData, seedTime: e.target.value ? parseInt(e.target.value, 10) : null })}
                    placeholder=""
                  />
                  <p className="text-xs text-muted-foreground">Empty uses download client default</p>
                </div>
              </div>
            </>
          )}

          {/* Priority */}
          <div className="space-y-2">
            <Label htmlFor="priority" className="text-amber-500">Indexer Priority</Label>
            <Input
              id="priority"
              type="number"
              min={1}
              max={50}
              value={formData.priority}
              onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value, 10) || 25 })}
            />
            <p className="text-xs text-muted-foreground">
              Indexer Priority from 1 (Highest) to 50 (Lowest). Default: 25. Used as tiebreaker for otherwise equal releases.
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
                testResult.success ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400" : "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400"
              }`}
            >
              {testResult.message}
            </div>
          )}

          {/* Actions */}
          <div className="flex justify-end gap-2 pt-4 border-t">
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

export function IndexerList() {
  const { data: indexers, isLoading, error } = useIndexers()
  const deleteIndexer = useDeleteIndexer()
  const createIndexer = useCreateIndexer()
  const updateIndexer = useUpdateIndexer()
  const [showTypeSelector, setShowTypeSelector] = useState(false)
  const [selectedType, setSelectedType] = useState<IndexerType | null>(null)
  const [selectedPreset, setSelectedPreset] = useState<IndexerPreset | undefined>(undefined)
  const [editingIndexer, setEditingIndexer] = useState<Indexer | null>(null)

  if (isLoading) {
    return <div className="text-muted-foreground">Loading indexers...</div>
  }

  if (error) {
    return <div className="text-destructive">Error loading indexers: {error.message}</div>
  }

  const handleSelectType = (type: IndexerType, preset?: IndexerPreset) => {
    setShowTypeSelector(false)
    setSelectedType(type)
    setSelectedPreset(preset)
  }

  const handleCreate = async (data: CreateIndexerInput) => {
    await createIndexer.mutateAsync(data)
    setSelectedType(null)
    setSelectedPreset(undefined)
  }

  const handleUpdate = async (data: CreateIndexerInput) => {
    if (!editingIndexer) return
    await updateIndexer.mutateAsync({ id: editingIndexer.id, data })
    setEditingIndexer(null)
  }

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this indexer?")) return
    await deleteIndexer.mutateAsync(id)
  }

  const handleCancelCreate = () => {
    setSelectedType(null)
    setSelectedPreset(undefined)
  }

  const isFormOpen = showTypeSelector || selectedType !== null || editingIndexer !== null

  return (
    <div className="space-y-6">
      {/* Add button with + icon like Radarr */}
      {!isFormOpen && (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
          {indexers?.map((indexer) => (
            <Card
              key={indexer.id}
              className="cursor-pointer hover:border-primary transition-colors"
              onClick={() => setEditingIndexer(indexer)}
            >
              <CardContent className="p-4">
                <div className="font-medium truncate">{indexer.name}</div>
                <div className="flex flex-wrap gap-1 mt-2">
                  {indexer.enableRss && <Badge variant="secondary" className="text-xs">RSS</Badge>}
                  {indexer.enableAutomaticSearch && <Badge variant="secondary" className="text-xs">Auto</Badge>}
                  {indexer.enableInteractiveSearch && <Badge variant="secondary" className="text-xs">Interactive</Badge>}
                </div>
                <div className="text-xs text-muted-foreground mt-2">Priority: {indexer.priority}</div>
                {!indexer.enabled && (
                  <Badge variant="outline" className="mt-2 text-xs">Disabled</Badge>
                )}
              </CardContent>
            </Card>
          ))}
          
          {/* Add button */}
          <Card
            className="cursor-pointer hover:border-primary transition-colors border-dashed"
            onClick={() => setShowTypeSelector(true)}
          >
            <CardContent className="p-4 flex items-center justify-center min-h-[100px]">
              <span className="text-4xl text-muted-foreground">+</span>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Type selector modal */}
      {showTypeSelector && (
        <IndexerTypeSelector
          onSelect={handleSelectType}
          onCancel={() => setShowTypeSelector(false)}
        />
      )}

      {/* Create form */}
      {selectedType && (
        <IndexerForm
          indexerType={selectedType}
          preset={selectedPreset}
          onSubmit={handleCreate}
          onCancel={handleCancelCreate}
          isLoading={createIndexer.isPending}
        />
      )}

      {/* Edit form */}
      {editingIndexer && (
        <div className="space-y-4">
          <IndexerForm
            indexer={editingIndexer}
            indexerType={editingIndexer.type}
            onSubmit={handleUpdate}
            onCancel={() => setEditingIndexer(null)}
            isLoading={updateIndexer.isPending}
          />
          <div className="flex justify-end">
            <Button
              variant="destructive"
              onClick={() => {
                handleDelete(editingIndexer.id)
                setEditingIndexer(null)
              }}
              disabled={deleteIndexer.isPending}
            >
              Delete Indexer
            </Button>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!indexers?.length && !isFormOpen && (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            No indexers configured. Click the + button to add an indexer.
          </CardContent>
        </Card>
      )}
    </div>
  )
}
