import { useState } from "react"
import type { DownloadClientPreset } from "../../components/download/download-client-presets"
import { Button, Card, CardContent, CardHeader, CardTitle, Input, Label } from "../../components/ui"
import { useTestDownloadClient } from "../../hooks/useDownload"
import type { CreateDownloadClientInput, DownloadClient, DownloadClientType } from "../../lib/api"

interface DownloadClientFormProps {
  client?: DownloadClient
  clientType: DownloadClientType
  preset?: DownloadClientPreset
  onSubmit: (data: CreateDownloadClientInput) => void
  onCancel: () => void
  isLoading?: boolean
}

export function DownloadClientForm({
  client,
  clientType,
  preset,
  onSubmit,
  onCancel,
  isLoading,
}: DownloadClientFormProps) {
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
  const title = client
    ? `Edit Download Client - ${client.name}`
    : `Add Download Client - ${preset?.name || clientType}`

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
              onChange={(e) =>
                setFormData({ ...formData, priority: parseInt(e.target.value, 10) || 1 })
              }
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
                testResult.success
                  ? "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
                  : "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400"
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
