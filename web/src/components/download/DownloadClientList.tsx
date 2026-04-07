import { useState } from "react"
import { useDownloadClients, useDeleteDownloadClient, useTestDownloadClient, useCreateDownloadClient, useUpdateDownloadClient } from "../../hooks/useDownload"
import type { DownloadClient, CreateDownloadClientInput, DownloadClientType } from "../../lib/api"
import { Button, Input, Label, Card, CardHeader, CardTitle, CardContent, Badge } from "../ui"

// Preset configurations for popular download clients
interface DownloadClientPreset {
  id: string
  name: string
  type: DownloadClientType
  host: string
  port: number
  description: string
}

const DOWNLOAD_CLIENT_PRESETS: { usenet: DownloadClientPreset[]; torrents: DownloadClientPreset[] } = {
  usenet: [
    {
      id: "sabnzbd",
      name: "SABnzbd",
      type: "sabnzbd",
      host: "localhost",
      port: 8080,
      description: "Popular Usenet downloader",
    },
  ],
  torrents: [
    {
      id: "qbittorrent",
      name: "qBittorrent",
      type: "qbittorrent",
      host: "localhost",
      port: 8080,
      description: "Feature-rich torrent client",
    },
    {
      id: "transmission",
      name: "Transmission",
      type: "transmission",
      host: "localhost",
      port: 9091,
      description: "Lightweight torrent client",
    },
    {
      id: "blackhole-nzb",
      name: "Blackhole (NZB)",
      type: "blackhole",
      host: "",
      port: 0,
      description: "Drop NZB files to folder",
    },
    {
      id: "blackhole-torrent",
      name: "Blackhole (Torrent)",
      type: "blackhole",
      host: "",
      port: 0,
      description: "Drop torrent files to folder",
    },
  ],
}

interface DownloadClientTypeSelectorProps {
  onSelect: (type: DownloadClientType, preset?: DownloadClientPreset) => void
  onCancel: () => void
}

function DownloadClientTypeSelector({ onSelect, onCancel }: DownloadClientTypeSelectorProps) {
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

interface DownloadClientFormProps {
  client?: DownloadClient
  clientType: DownloadClientType
  preset?: DownloadClientPreset
  onSubmit: (data: CreateDownloadClientInput) => void
  onCancel: () => void
  isLoading?: boolean
}

function DownloadClientForm({ client, clientType, preset, onSubmit, onCancel, isLoading }: DownloadClientFormProps) {
  const [formData, setFormData] = useState<CreateDownloadClientInput>({
    name: client?.name ?? preset?.name ?? "",
    type: client?.type ?? clientType,
    host: client?.host ?? preset?.host ?? "localhost",
    port: client?.port ?? preset?.port ?? 8080,
    useTls: client?.useTls ?? false,
    username: client?.username ?? "",
    password: client?.password ?? "",
    apiKey: client?.apiKey ?? "",
    category: client?.category ?? "books",
    recentPriority: client?.recentPriority ?? 0,
    olderPriority: client?.olderPriority ?? 0,
    removeCompletedAfter: client?.removeCompletedAfter ?? 0,
    enabled: client?.enabled ?? true,
    priority: client?.priority ?? 1,
    nzbFolder: client?.nzbFolder ?? "",
    torrentFolder: client?.torrentFolder ?? "",
    watchFolder: client?.watchFolder ?? "",
  })
  
  const testClient = useTestDownloadClient()
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)

  const handleTest = async () => {
    setTestResult(null)
    try {
      const result = await testClient.mutateAsync(formData)
      setTestResult(result)
    } catch {
      setTestResult({ success: false, message: "Connection failed" })
    }
  }

  const isBlackhole = formData.type === "blackhole"
  const isSabnzbd = formData.type === "sabnzbd"
  const title = client ? `Edit Download Client - ${client.name}` : `Add Download Client - ${preset?.name || clientType}`

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
              placeholder="My Download Client"
              required
            />
          </div>

          {/* Host and Port (not for blackhole) */}
          {!isBlackhole && (
            <div className="grid grid-cols-3 gap-4">
              <div className="col-span-2 space-y-2">
                <Label htmlFor="host">Host</Label>
                <Input
                  id="host"
                  value={formData.host}
                  onChange={(e) => setFormData({ ...formData, host: e.target.value })}
                  placeholder="localhost"
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="port">Port</Label>
                <Input
                  id="port"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  value={formData.port || ""}
                  onChange={(e) => {
                    const val = e.target.value.replace(/\D/g, "")
                    setFormData({ ...formData, port: val ? parseInt(val, 10) : 0 })
                  }}
                  placeholder="8080"
                  required
                />
              </div>
            </div>
          )}

          {/* TLS */}
          {!isBlackhole && (
            <div className="flex items-center gap-2">
              <input
                id="useTls"
                type="checkbox"
                checked={formData.useTls}
                onChange={(e) => setFormData({ ...formData, useTls: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300"
              />
              <Label htmlFor="useTls">Use TLS (HTTPS)</Label>
            </div>
          )}

          {/* API Key (SABnzbd) */}
          {isSabnzbd && (
            <div className="space-y-2">
              <Label htmlFor="apiKey">API Key</Label>
              <Input
                id="apiKey"
                type="password"
                value={formData.apiKey}
                onChange={(e) => setFormData({ ...formData, apiKey: e.target.value })}
                placeholder="Your SABnzbd API key"
              />
            </div>
          )}

          {/* Username/Password (qBittorrent, Transmission) */}
          {!isBlackhole && !isSabnzbd && (
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  placeholder="admin"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                />
              </div>
            </div>
          )}

          {/* Blackhole folders */}
          {isBlackhole && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="nzbFolder">NZB Folder</Label>
                <Input
                  id="nzbFolder"
                  value={formData.nzbFolder}
                  onChange={(e) => setFormData({ ...formData, nzbFolder: e.target.value })}
                  placeholder="/path/to/nzb/watch"
                />
                <p className="text-xs text-muted-foreground">Folder to drop NZB files</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="torrentFolder">Torrent Folder</Label>
                <Input
                  id="torrentFolder"
                  value={formData.torrentFolder}
                  onChange={(e) => setFormData({ ...formData, torrentFolder: e.target.value })}
                  placeholder="/path/to/torrent/watch"
                />
                <p className="text-xs text-muted-foreground">Folder to drop torrent/magnet files</p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="watchFolder">Watch Folder</Label>
                <Input
                  id="watchFolder"
                  value={formData.watchFolder}
                  onChange={(e) => setFormData({ ...formData, watchFolder: e.target.value })}
                  placeholder="/path/to/watch"
                />
                <p className="text-xs text-muted-foreground">General watch folder (fallback)</p>
              </div>
            </div>
          )}

          {/* Category */}
          {!isBlackhole && (
            <div className="space-y-2">
              <Label htmlFor="category">Category</Label>
              <Input
                id="category"
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                placeholder="books"
              />
              <p className="text-xs text-muted-foreground">Category to use in download client</p>
            </div>
          )}

          {/* Priority */}
          <div className="space-y-2">
            <Label htmlFor="priority">Client Priority</Label>
            <Input
              id="priority"
              type="number"
              min={1}
              max={50}
              value={formData.priority}
              onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value, 10) || 1 })}
            />
            <p className="text-xs text-muted-foreground">
              Priority from 1 (highest) to 50 (lowest). Lower priority clients are used first.
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
              disabled={testClient.isPending}
            >
              {testClient.isPending ? "Testing..." : "Test"}
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

export function DownloadClientList() {
  const { data: clients, isLoading, error } = useDownloadClients()
  const deleteClient = useDeleteDownloadClient()
  const createClient = useCreateDownloadClient()
  const updateClient = useUpdateDownloadClient()
  const [showTypeSelector, setShowTypeSelector] = useState(false)
  const [selectedType, setSelectedType] = useState<DownloadClientType | null>(null)
  const [selectedPreset, setSelectedPreset] = useState<DownloadClientPreset | undefined>(undefined)
  const [editingClient, setEditingClient] = useState<DownloadClient | null>(null)

  if (isLoading) {
    return <div className="text-muted-foreground">Loading download clients...</div>
  }

  if (error) {
    return <div className="text-destructive">Error loading download clients: {error.message}</div>
  }

  const handleSelectType = (type: DownloadClientType, preset?: DownloadClientPreset) => {
    setShowTypeSelector(false)
    setSelectedType(type)
    setSelectedPreset(preset)
  }

  const handleCreate = async (data: CreateDownloadClientInput) => {
    await createClient.mutateAsync(data)
    setSelectedType(null)
    setSelectedPreset(undefined)
  }

  const handleUpdate = async (data: CreateDownloadClientInput) => {
    if (!editingClient) return
    await updateClient.mutateAsync({ id: editingClient.id, data })
    setEditingClient(null)
  }

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this download client?")) return
    await deleteClient.mutateAsync(id)
  }

  const handleCancelCreate = () => {
    setSelectedType(null)
    setSelectedPreset(undefined)
  }

  const isFormOpen = showTypeSelector || selectedType !== null || editingClient !== null

  const getClientTypeLabel = (type: DownloadClientType) => {
    switch (type) {
      case "sabnzbd": return "SABnzbd"
      case "qbittorrent": return "qBittorrent"
      case "transmission": return "Transmission"
      case "blackhole": return "Blackhole"
      default: return type
    }
  }

  return (
    <div className="space-y-6">
      {/* Client cards */}
      {!isFormOpen && (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
          {clients?.map((client) => (
            <Card
              key={client.id}
              className="cursor-pointer hover:border-primary transition-colors"
              onClick={() => setEditingClient(client)}
            >
              <CardContent className="p-4">
                <div className="font-medium truncate">{client.name}</div>
                <Badge variant="secondary" className="mt-2 text-xs">
                  {getClientTypeLabel(client.type)}
                </Badge>
                {!client.enabled && (
                  <Badge variant="outline" className="ml-2 text-xs">Disabled</Badge>
                )}
                <div className="text-xs text-muted-foreground mt-2">
                  {client.type === "blackhole" 
                    ? "File drop" 
                    : `${client.host}:${client.port}`}
                </div>
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

      {/* Type selector */}
      {showTypeSelector && (
        <DownloadClientTypeSelector
          onSelect={handleSelectType}
          onCancel={() => setShowTypeSelector(false)}
        />
      )}

      {/* Create form */}
      {selectedType && (
        <DownloadClientForm
          clientType={selectedType}
          preset={selectedPreset}
          onSubmit={handleCreate}
          onCancel={handleCancelCreate}
          isLoading={createClient.isPending}
        />
      )}

      {/* Edit form */}
      {editingClient && (
        <div className="space-y-4">
          <DownloadClientForm
            client={editingClient}
            clientType={editingClient.type}
            onSubmit={handleUpdate}
            onCancel={() => setEditingClient(null)}
            isLoading={updateClient.isPending}
          />
          <div className="flex justify-end">
            <Button
              variant="destructive"
              onClick={() => {
                handleDelete(editingClient.id)
                setEditingClient(null)
              }}
              disabled={deleteClient.isPending}
            >
              Delete Download Client
            </Button>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!clients?.length && !isFormOpen && (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            <p>No download clients configured.</p>
            <p className="text-sm">Click the + button to add one.</p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
