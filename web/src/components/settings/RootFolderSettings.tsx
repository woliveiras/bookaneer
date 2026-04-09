import { useState } from "react"
import { useRootFolders, useCreateRootFolder, useUpdateRootFolder, useDeleteRootFolder } from "../../hooks/useRootFolders"
import type { RootFolder, CreateRootFolderInput, UpdateRootFolderInput } from "../../lib/api"
import { Button, Input, Label, Card, CardHeader, CardTitle, CardContent, Badge } from "../ui"

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B"
  const k = 1024
  const sizes = ["B", "KB", "MB", "GB", "TB"]
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`
}

interface RootFolderFormProps {
  folder?: RootFolder
  onSubmit: (data: CreateRootFolderInput | UpdateRootFolderInput) => Promise<void>
  onCancel: () => void
  isLoading: boolean
}

function RootFolderForm({ folder, onSubmit, onCancel, isLoading }: RootFolderFormProps) {
  const [formData, setFormData] = useState<CreateRootFolderInput>({
    path: folder?.path || "",
    name: folder?.name || "",
  })
  const [moveFiles, setMoveFiles] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isEditing = !!folder
  const pathChanged = isEditing && formData.path !== folder.path

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    
    if (!formData.path.trim()) {
      setError("Path is required")
      return
    }
    if (!formData.name.trim()) {
      setError("Name is required")
      return
    }

    try {
      if (isEditing && pathChanged) {
        await onSubmit({ ...formData, moveFiles })
      } else {
        await onSubmit(formData)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save root folder")
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{folder ? "Edit Root Folder" : "Add Root Folder"}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="path">Path *</Label>
            <Input
              id="path"
              type="text"
              value={formData.path}
              onChange={(e) => setFormData({ ...formData, path: e.target.value })}
              placeholder="/path/to/books"
              required
            />
            <p className="text-xs text-muted-foreground">
              The folder where your book library is stored. Must be an existing directory.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="My Books"
              required
            />
            <p className="text-xs text-muted-foreground">
              A friendly name for this folder.
            </p>
          </div>

          {/* Move files option - only show when editing and path changed */}
          {isEditing && pathChanged && (
            <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-md space-y-3">
              <div className="flex items-start gap-3">
                <input
                  type="checkbox"
                  id="moveFiles"
                  checked={moveFiles}
                  onChange={(e) => setMoveFiles(e.target.checked)}
                  className="mt-1 h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                />
                <div>
                  <Label htmlFor="moveFiles" className="text-amber-600 font-medium cursor-pointer">
                    Move existing files to new location
                  </Label>
                  <p className="text-xs text-muted-foreground mt-1">
                    When enabled, all author folders and book files will be moved from the old path to the new path. 
                    This operation may take some time depending on library size.
                  </p>
                </div>
              </div>
              {!moveFiles && (
                <p className="text-xs text-amber-600">
                  ⚠️ Without this option, existing files stay in place and you'll need to move them manually.
                </p>
              )}
            </div>
          )}

          <div className="flex gap-3 pt-2">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading 
                ? (pathChanged && moveFiles ? "Moving files..." : "Saving...") 
                : (pathChanged && moveFiles ? "Save & Move Files" : "Save")
              }
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

export function RootFolderList() {
  const { data: folders, isLoading, error } = useRootFolders()
  const createFolder = useCreateRootFolder()
  const updateFolder = useUpdateRootFolder()
  const deleteFolder = useDeleteRootFolder()
  const [showAddForm, setShowAddForm] = useState(false)
  const [editingFolder, setEditingFolder] = useState<RootFolder | null>(null)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  if (isLoading) {
    return <div className="text-muted-foreground">Loading root folders...</div>
  }

  if (error) {
    return <div className="text-destructive">Error loading root folders: {error.message}</div>
  }

  const handleCreate = async (data: CreateRootFolderInput | UpdateRootFolderInput) => {
    await createFolder.mutateAsync(data as CreateRootFolderInput)
    setShowAddForm(false)
  }

  const handleUpdate = async (data: CreateRootFolderInput | UpdateRootFolderInput) => {
    if (!editingFolder) return
    await updateFolder.mutateAsync({ id: editingFolder.id, data: data as UpdateRootFolderInput })
    setEditingFolder(null)
  }

  const handleDelete = async () => {
    if (!editingFolder) return
    await deleteFolder.mutateAsync(editingFolder.id)
    setShowDeleteConfirm(false)
    setEditingFolder(null)
  }

  const isFormOpen = showAddForm || editingFolder !== null

  return (
    <div className="space-y-6">
      {/* Folder cards */}
      {!isFormOpen && (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {folders?.map((folder) => (
            <Card
              key={folder.id}
              className="cursor-pointer hover:border-primary transition-colors"
              onClick={() => setEditingFolder(folder)}
            >
              <CardContent className="p-4">
                <div className="font-medium truncate">{folder.name}</div>
                <div className="text-xs text-muted-foreground mt-1 font-mono truncate" title={folder.path}>
                  {folder.path}
                </div>
                <div className="flex gap-2 mt-2">
                  {folder.accessible ? (
                    <Badge variant="secondary" className="text-xs">Accessible</Badge>
                  ) : (
                    <Badge variant="destructive" className="text-xs">Not Accessible</Badge>
                  )}
                  {folder.authorCount !== undefined && folder.authorCount > 0 && (
                    <Badge variant="outline" className="text-xs">
                      {folder.authorCount} authors
                    </Badge>
                  )}
                </div>
                {folder.freeSpace !== undefined && folder.totalSpace !== undefined && folder.totalSpace > 0 && (
                  <div className="mt-2">
                    <div className="flex justify-between text-xs text-muted-foreground mb-1">
                      <span>{formatBytes(folder.freeSpace)} free</span>
                      <span>{formatBytes(folder.totalSpace)} total</span>
                    </div>
                    <div className="h-1.5 bg-muted rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-primary rounded-full"
                        style={{ width: `${((folder.totalSpace - folder.freeSpace) / folder.totalSpace) * 100}%` }}
                      />
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
          
          {/* Add button */}
          <Card
            className="cursor-pointer hover:border-primary transition-colors border-dashed"
            onClick={() => setShowAddForm(true)}
          >
            <CardContent className="p-4 flex items-center justify-center min-h-[120px]">
              <div className="text-center">
                <span className="text-4xl text-muted-foreground">+</span>
                <p className="text-sm text-muted-foreground mt-2">Add Root Folder</p>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Add form */}
      {showAddForm && (
        <RootFolderForm
          onSubmit={handleCreate}
          onCancel={() => setShowAddForm(false)}
          isLoading={createFolder.isPending}
        />
      )}

      {/* Edit form */}
      {editingFolder && (
        <div className="space-y-4">
          <RootFolderForm
            folder={editingFolder}
            onSubmit={handleUpdate}
            onCancel={() => setEditingFolder(null)}
            isLoading={updateFolder.isPending}
          />
          <div className="flex justify-end">
            <Button
              variant="destructive"
              onClick={() => setShowDeleteConfirm(true)}
              disabled={deleteFolder.isPending}
            >
              Delete Root Folder
            </Button>
          </div>
        </div>
      )}

      {/* Delete confirmation dialog */}
      {showDeleteConfirm && editingFolder && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background p-6 rounded-lg shadow-lg max-w-md w-full mx-4 border">
            <h3 className="text-lg font-semibold mb-2">Delete Root Folder?</h3>
            <p className="text-muted-foreground mb-4">
              Are you sure you want to delete "{editingFolder.name}"? This will only remove the folder from Bookaneer, not delete any files on disk.
            </p>
            <div className="flex gap-2 justify-end">
              <Button variant="outline" onClick={() => setShowDeleteConfirm(false)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteFolder.isPending}
              >
                {deleteFolder.isPending ? "Deleting..." : "Delete"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!folders?.length && !isFormOpen && (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            <p className="text-lg font-medium mb-2">No root folders configured</p>
            <p className="text-sm">
              Add a root folder to store your book library. This is where downloaded books will be saved.
            </p>
            <Button className="mt-4" onClick={() => setShowAddForm(true)}>
              Add Root Folder
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
