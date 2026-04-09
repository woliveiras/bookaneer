import { useState } from "react"
import { DownloadClientTypeSelector } from "../../components/download/DownloadClientTypeSelector"
import type { DownloadClientPreset } from "../../components/download/download-client-presets"
import { Badge, Button, Card, CardContent } from "../../components/ui"
import {
  useCreateDownloadClient,
  useDeleteDownloadClient,
  useDownloadClients,
  useUpdateDownloadClient,
} from "../../hooks/useDownload"
import type { CreateDownloadClientInput, DownloadClient, DownloadClientType } from "../../lib/api"
import { DownloadClientForm } from "./DownloadClientForm"

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
      case "sabnzbd":
        return "SABnzbd"
      case "qbittorrent":
        return "qBittorrent"
      case "transmission":
        return "Transmission"
      case "blackhole":
        return "Blackhole"
      default:
        return type
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
                  <Badge variant="outline" className="ml-2 text-xs">
                    Disabled
                  </Badge>
                )}
                <div className="text-xs text-muted-foreground mt-2">
                  {client.type === "blackhole" ? "File drop" : `${client.host}:${client.port}`}
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
