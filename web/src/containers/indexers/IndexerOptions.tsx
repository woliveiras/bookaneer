import { useEffect, useState } from "react"
import { Button, Input, Label } from "../../components/ui"
import { useIndexerOptions, useUpdateIndexerOptions } from "../../hooks/useIndexers"

export function IndexerOptions() {
  const { data: options, isLoading, error } = useIndexerOptions()
  const updateOptions = useUpdateIndexerOptions()
  const [formData, setFormData] = useState({
    minimumAge: 0,
    retention: 0,
    maximumSize: 0,
    rssSyncInterval: 30,
    preferIndexerFlags: false,
    availabilityDelay: 0,
  })
  const [isDirty, setIsDirty] = useState(false)

  useEffect(() => {
    if (options) {
      setFormData({
        minimumAge: options.minimumAge,
        retention: options.retention,
        maximumSize: options.maximumSize,
        rssSyncInterval: options.rssSyncInterval,
        preferIndexerFlags: options.preferIndexerFlags,
        availabilityDelay: options.availabilityDelay,
      })
    }
  }, [options])

  const handleChange = (field: keyof typeof formData, value: number | boolean) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
    setIsDirty(true)
  }

  const handleSave = async () => {
    await updateOptions.mutateAsync(formData)
    setIsDirty(false)
  }

  if (isLoading) {
    return <div className="text-muted-foreground">Loading options...</div>
  }

  if (error) {
    return <div className="text-destructive">Error loading options: {error.message}</div>
  }

  return (
    <div className="space-y-6">
      <h4 className="text-md font-medium text-muted-foreground">Indexer Options</h4>
      <div className="grid gap-6 md:grid-cols-2">
        {/* Minimum Age */}
        <div className="space-y-2">
          <Label htmlFor="minimumAge">Minimum Age (minutes)</Label>
          <div className="flex items-center gap-2">
            <Input
              id="minimumAge"
              type="number"
              min={0}
              value={formData.minimumAge}
              onChange={(e) => handleChange("minimumAge", parseInt(e.target.value, 10) || 0)}
              className="flex-1"
            />
            <span className="text-sm text-muted-foreground w-16">minutes</span>
          </div>
          <p className="text-xs text-muted-foreground">
            Usenet only: Minimum age of releases before they are grabbed. Use to give new releases
            time to propagate.
          </p>
        </div>

        {/* Retention */}
        <div className="space-y-2">
          <Label htmlFor="retention">Retention (days)</Label>
          <div className="flex items-center gap-2">
            <Input
              id="retention"
              type="number"
              min={0}
              value={formData.retention}
              onChange={(e) => handleChange("retention", parseInt(e.target.value, 10) || 0)}
              className="flex-1"
            />
            <span className="text-sm text-muted-foreground w-16">days</span>
          </div>
          <p className="text-xs text-muted-foreground">
            Usenet only: Set to zero for unlimited retention.
          </p>
        </div>

        {/* Maximum Size */}
        <div className="space-y-2">
          <Label htmlFor="maximumSize">Maximum Size (MB)</Label>
          <div className="flex items-center gap-2">
            <Input
              id="maximumSize"
              type="number"
              min={0}
              value={formData.maximumSize}
              onChange={(e) => handleChange("maximumSize", parseInt(e.target.value, 10) || 0)}
              className="flex-1"
            />
            <span className="text-sm text-muted-foreground w-16">MB</span>
          </div>
          <p className="text-xs text-muted-foreground">
            Maximum size for a release to be grabbed. Set to zero for unlimited.
          </p>
        </div>

        {/* RSS Sync Interval */}
        <div className="space-y-2">
          <Label htmlFor="rssSyncInterval" className="text-amber-500">
            RSS Sync Interval
          </Label>
          <div className="flex items-center gap-2">
            <Input
              id="rssSyncInterval"
              type="number"
              min={0}
              value={formData.rssSyncInterval}
              onChange={(e) => handleChange("rssSyncInterval", parseInt(e.target.value, 10) || 0)}
              className="flex-1"
            />
            <span className="text-sm text-muted-foreground w-16">minutes</span>
          </div>
          <p className="text-xs text-muted-foreground">
            Interval in minutes. Set to zero to disable automatic release grabbing.
          </p>
          <p className="text-xs text-amber-500">
            This will apply to all indexers. Please follow the rules set forth by them.
          </p>
        </div>

        {/* Availability Delay */}
        <div className="space-y-2">
          <Label htmlFor="availabilityDelay">Availability Delay (days)</Label>
          <div className="flex items-center gap-2">
            <Input
              id="availabilityDelay"
              type="number"
              min={0}
              value={formData.availabilityDelay}
              onChange={(e) => handleChange("availabilityDelay", parseInt(e.target.value, 10) || 0)}
              className="flex-1"
            />
            <span className="text-sm text-muted-foreground w-16">days</span>
          </div>
          <p className="text-xs text-muted-foreground">
            Amount of time before or after release date to search for books.
          </p>
        </div>

        {/* Prefer Indexer Flags */}
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <input
              id="preferIndexerFlags"
              type="checkbox"
              checked={formData.preferIndexerFlags}
              onChange={(e) => handleChange("preferIndexerFlags", e.target.checked)}
              className="h-4 w-4 rounded border-gray-300"
            />
            <Label htmlFor="preferIndexerFlags">Prefer Indexer Flags</Label>
          </div>
          <p className="text-xs text-muted-foreground">
            Prioritize releases with special flags (e.g., internal, scene).
          </p>
        </div>
      </div>

      <div className="flex justify-end pt-4 border-t">
        <Button onClick={handleSave} disabled={!isDirty || updateOptions.isPending}>
          {updateOptions.isPending ? "Saving..." : "Save Options"}
        </Button>
      </div>
    </div>
  )
}
