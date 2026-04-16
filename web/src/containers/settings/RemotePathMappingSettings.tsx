import { Trash2 } from "lucide-react"
import { useState } from "react"
import { Button } from "../../components/ui/Button"
import { Card, CardContent } from "../../components/ui/Card"
import { Input } from "../../components/ui/Input"
import { Label } from "../../components/ui/Label"
import {
  useCreateRemotePathMapping,
  useDeleteRemotePathMapping,
  useRemotePathMappings,
} from "../../hooks/useRemotePathMappings"

export function RemotePathMappingSettings() {
  const { data: mappings } = useRemotePathMappings()
  const createMapping = useCreateRemotePathMapping()
  const deleteMapping = useDeleteRemotePathMapping()

  const [showForm, setShowForm] = useState(false)
  const [host, setHost] = useState("")
  const [remotePath, setRemotePath] = useState("")
  const [localPath, setLocalPath] = useState("")

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!remotePath || !localPath) return

    createMapping.mutate(
      { host, remotePath, localPath },
      {
        onSuccess: () => {
          setHost("")
          setRemotePath("")
          setLocalPath("")
          setShowForm(false)
        },
      },
    )
  }

  function handleDelete(id: number) {
    deleteMapping.mutate(id)
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        If your download client runs in a different container or machine, its file paths may differ
        from Bookaneer's. Add mappings to translate remote paths to local paths.
      </p>

      {mappings && mappings.length > 0 && (
        <div className="space-y-2">
          {mappings.map((m) => (
            <Card key={m.id}>
              <CardContent className="p-3 flex items-center gap-4">
                <div className="flex-1 min-w-0 grid grid-cols-1 sm:grid-cols-3 gap-2 text-sm">
                  <div>
                    <span className="text-muted-foreground">Host: </span>
                    <span className="font-mono">{m.host || "—"}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Remote: </span>
                    <span className="font-mono break-all">{m.remotePath}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Local: </span>
                    <span className="font-mono break-all">{m.localPath}</span>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleDelete(m.id)}
                  disabled={deleteMapping.isPending}
                  aria-label={`Delete mapping for ${m.remotePath}`}
                >
                  <Trash2 className="w-4 h-4 text-destructive" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {showForm ? (
        <form onSubmit={handleSubmit} className="space-y-3 border rounded-lg p-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div>
              <Label htmlFor="rpm-host">Host (optional)</Label>
              <Input
                id="rpm-host"
                value={host}
                onChange={(e) => setHost(e.target.value)}
                placeholder="e.g. qbittorrent"
              />
            </div>
            <div>
              <Label htmlFor="rpm-remote">Remote Path</Label>
              <Input
                id="rpm-remote"
                value={remotePath}
                onChange={(e) => setRemotePath(e.target.value)}
                placeholder="/data/downloads/"
                required
              />
            </div>
            <div>
              <Label htmlFor="rpm-local">Local Path</Label>
              <Input
                id="rpm-local"
                value={localPath}
                onChange={(e) => setLocalPath(e.target.value)}
                placeholder="/media/downloads/"
                required
              />
            </div>
          </div>
          <div className="flex gap-2">
            <Button type="submit" disabled={createMapping.isPending}>
              {createMapping.isPending ? "Saving..." : "Save"}
            </Button>
            <Button type="button" variant="outline" onClick={() => setShowForm(false)}>
              Cancel
            </Button>
          </div>
        </form>
      ) : (
        <Button variant="outline" onClick={() => setShowForm(true)}>
          Add Path Mapping
        </Button>
      )}
    </div>
  )
}
