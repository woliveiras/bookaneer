import { useEffect, useState } from "react"
import { Button } from "../../components/ui/Button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card"
import { Input } from "../../components/ui/Input"
import { Label } from "../../components/ui/Label"
import {
  useNamingPreview,
  useNamingSettings,
  usePreviewRenameAll,
  useRenameAll,
  useUpdateNamingSettings,
} from "../../hooks/useNaming"
import type { NamingPreviewInput } from "../../lib/types"

const TEMPLATE_VARIABLES = [
  { token: "$Author", description: "Author name" },
  { token: "$SortAuthor", description: "Author sort name" },
  { token: "$Title", description: "Book title" },
  { token: "$Series", description: "Series name" },
  { token: "$SeriesName", description: "Series name (alias)" },
  { token: "$SeriesPosition", description: "Position in series" },
  { token: "$Year", description: "Publication year" },
  { token: "$Format", description: "File format (epub, pdf, etc.)" },
  { token: "$Quality", description: "Quality profile" },
  { token: "$OriginalName", description: "Original file name" },
]

const COLON_OPTIONS = [
  { value: "dash", label: "Replace with dash (: → -)" },
  { value: "space", label: "Replace with space" },
  { value: "delete", label: "Remove colons" },
]

export function NamingSettings() {
  const { data: settings, isLoading, error } = useNamingSettings()
  const updateSettings = useUpdateNamingSettings()
  const renameAll = useRenameAll()
  const previewRenameAll = usePreviewRenameAll()

  const [authorFormat, setAuthorFormat] = useState("")
  const [bookFormat, setBookFormat] = useState("")
  const [replaceSpaces, setReplaceSpaces] = useState(false)
  const [colonReplacement, setColonReplacement] = useState("dash")
  const [dirty, setDirty] = useState(false)
  const [showRenameConfirm, setShowRenameConfirm] = useState(false)

  useEffect(() => {
    if (settings) {
      setAuthorFormat(settings.authorFolderFormat)
      setBookFormat(settings.bookFileFormat)
      setReplaceSpaces(settings.replaceSpaces)
      setColonReplacement(settings.colonReplacement)
    }
  }, [settings])

  const previewInput: NamingPreviewInput | null =
    authorFormat && bookFormat
      ? {
          authorFolderFormat: authorFormat,
          bookFileFormat: bookFormat,
          replaceSpaces,
          colonReplacement,
        }
      : null

  const { data: preview } = useNamingPreview(previewInput)

  function handleFieldChange<T>(setter: (v: T) => void, value: T) {
    setter(value)
    setDirty(true)
  }

  async function handleSave() {
    await updateSettings.mutateAsync({
      authorFolderFormat: authorFormat,
      bookFileFormat: bookFormat,
      replaceSpaces,
      colonReplacement,
    })
    setDirty(false)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 text-destructive bg-destructive/10 rounded-md">
        Failed to load naming settings: {error.message}
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Templates */}
      <Card>
        <CardHeader>
          <CardTitle>File Naming</CardTitle>
          <CardDescription>
            Configure how book files and folders are named in your library
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label htmlFor="author-format">Author Folder Format</Label>
            <Input
              id="author-format"
              value={authorFormat}
              onChange={(e) => handleFieldChange(setAuthorFormat, e.target.value)}
              placeholder="$Author"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Template for the author folder name
            </p>
          </div>

          <div>
            <Label htmlFor="book-format">Book File Format</Label>
            <Input
              id="book-format"
              value={bookFormat}
              onChange={(e) => handleFieldChange(setBookFormat, e.target.value)}
              placeholder="$Author - $Title"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Template for the book file name (extension is added automatically). Use {"{ }"} for
              conditional blocks, e.g. {"$Title{ ($SeriesName #$SeriesPosition)}"}
            </p>
          </div>

          <div className="flex items-center gap-2">
            <input
              id="replace-spaces"
              type="checkbox"
              checked={replaceSpaces}
              onChange={(e) => handleFieldChange(setReplaceSpaces, e.target.checked)}
              className="h-4 w-4 rounded border-border"
            />
            <Label htmlFor="replace-spaces">Replace spaces with dots</Label>
          </div>

          <div>
            <Label htmlFor="colon-replacement">Colon Replacement</Label>
            <select
              id="colon-replacement"
              value={colonReplacement}
              onChange={(e) => handleFieldChange(setColonReplacement, e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            >
              {COLON_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          {/* Preview */}
          {preview && (
            <div className="bg-muted rounded-md p-3 space-y-1">
              <p className="text-xs font-medium text-muted-foreground">Preview (sample book):</p>
              <p className="text-sm font-mono">{preview.relativePath}</p>
            </div>
          )}

          <div className="flex gap-2 pt-2">
            <Button onClick={handleSave} disabled={!dirty || updateSettings.isPending}>
              {updateSettings.isPending ? "Saving..." : "Save"}
            </Button>
            {updateSettings.isSuccess && !dirty && (
              <span className="text-sm text-green-500 self-center">Saved!</span>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Available Variables Reference */}
      <Card>
        <CardHeader>
          <CardTitle>Available Variables</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-x-6 gap-y-1 text-sm">
            {TEMPLATE_VARIABLES.map((v) => (
              <div key={v.token} className="flex gap-2">
                <code className="text-primary font-mono">{v.token}</code>
                <span className="text-muted-foreground">— {v.description}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Batch Rename */}
      <Card>
        <CardHeader>
          <CardTitle>Rename Library</CardTitle>
          <CardDescription>
            Rename all existing files in your library to match the current naming settings
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => previewRenameAll.mutate()}
              disabled={previewRenameAll.isPending}
            >
              {previewRenameAll.isPending ? "Previewing..." : "Preview Rename"}
            </Button>
            <Button
              variant="outline"
              onClick={() => setShowRenameConfirm(true)}
              disabled={renameAll.isPending}
            >
              {renameAll.isPending ? "Renaming..." : "Rename All Files"}
            </Button>
          </div>

          {/* Preview results */}
          {previewRenameAll.data && (
            <div className="bg-muted rounded-md p-3 space-y-2">
              <p className="text-sm font-medium">
                Preview: {previewRenameAll.data.renamed} files would be renamed,{" "}
                {previewRenameAll.data.skipped} already correct
              </p>
              {previewRenameAll.data.files && previewRenameAll.data.files.length > 0 && (
                <div className="max-h-48 overflow-y-auto space-y-1">
                  {previewRenameAll.data.files.map((f) => (
                    <div
                      key={`${f.oldPath}-${f.newPath}`}
                      className="text-xs font-mono space-y-0.5"
                    >
                      <p className="text-red-400">- {f.oldPath}</p>
                      <p className="text-green-400">+ {f.newPath}</p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Rename results */}
          {renameAll.data && (
            <div className="bg-muted rounded-md p-3 space-y-1">
              <p className="text-sm font-medium">
                Renamed: {renameAll.data.renamed} / {renameAll.data.total} ({renameAll.data.skipped}{" "}
                skipped)
              </p>
              {renameAll.data.errors && renameAll.data.errors.length > 0 && (
                <div className="text-xs text-destructive space-y-0.5">
                  {renameAll.data.errors.map((e) => (
                    <p key={e}>{e}</p>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Rename confirmation dialog */}
          {showRenameConfirm && (
            <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-md p-4 space-y-3">
              <p className="text-sm font-medium">
                Are you sure? This will rename all files in your library.
              </p>
              <div className="flex gap-2">
                <Button
                  variant="default"
                  onClick={() => {
                    renameAll.mutate()
                    setShowRenameConfirm(false)
                  }}
                >
                  Yes, Rename All
                </Button>
                <Button variant="outline" onClick={() => setShowRenameConfirm(false)}>
                  Cancel
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
