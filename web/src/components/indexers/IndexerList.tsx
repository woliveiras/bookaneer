import { useState } from "react"
import { useIndexers, useDeleteIndexer, useCreateIndexer, useUpdateIndexer } from "../../hooks/useIndexers"
import type { Indexer, CreateIndexerInput } from "../../lib/api"
import { Button, Card, CardContent, Badge } from "../ui"
import type { IndexerType, IndexerPreset } from "./indexer-presets"
import { IndexerTypeSelector } from "./IndexerTypeSelector"
import { IndexerForm } from "./IndexerForm"

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
