import { useCallback } from "react"
import { useReadingProgress, useSaveProgress } from "../../hooks/useReader"

/**
 * Manages reading progress: loads the last-saved position from the API
 * and provides a stable callback to persist progress as the reader relocates.
 */
export function useReaderProgress(bookFileId: number) {
  const { data: savedProgress } = useReadingProgress(bookFileId)
  const saveProgressMutation = useSaveProgress(bookFileId)

  const saveProgress = useCallback(
    (cfi: string, percentage: number) => {
      saveProgressMutation.mutate({ position: cfi, percentage })
    },
    [saveProgressMutation],
  )

  return { savedProgress, saveProgress }
}
