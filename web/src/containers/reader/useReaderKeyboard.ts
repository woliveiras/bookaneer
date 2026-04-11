import { useEffect } from "react"
import type { FoliateView } from "../../components/reader/readerConfig"

interface UseReaderKeyboardProps {
  viewRef: React.RefObject<FoliateView | null>
  readerReady: boolean
  showSettings: boolean
  showToc: boolean
  showBookmarks: boolean
  setShowSettings: React.Dispatch<React.SetStateAction<boolean>>
  setShowToc: React.Dispatch<React.SetStateAction<boolean>>
  setShowBookmarks: React.Dispatch<React.SetStateAction<boolean>>
  onPrev: () => void
  onNext: () => void
  onClose: () => void
}

export function useReaderKeyboard({
  viewRef,
  readerReady,
  showSettings,
  showToc,
  showBookmarks,
  setShowSettings,
  setShowToc,
  setShowBookmarks,
  onPrev,
  onNext,
  onClose,
}: UseReaderKeyboardProps) {
  useEffect(() => {
    if (!readerReady) {
      return
    }

    const handledKey = "__bookaneerReaderHandled"

    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e as KeyboardEvent & Record<string, boolean>)[handledKey]) {
        return
      }

      const code = e.code
      const target = e.target as HTMLElement | null
      const tagName = target?.tagName?.toLowerCase()
      const isEditable =
        target?.isContentEditable ||
        tagName === "input" ||
        tagName === "textarea" ||
        tagName === "select"

      if (isEditable) {
        return
      }

      // Global shortcuts should work even with panels open.
      if (e.key === "Escape") {
        ;(e as KeyboardEvent & Record<string, boolean>)[handledKey] = true
        e.preventDefault()
        if (showSettings || showToc || showBookmarks) {
          setShowSettings(false)
          setShowToc(false)
          setShowBookmarks(false)
        } else {
          onClose()
        }
        return
      }

      if (code === "KeyT" || e.key === "t" || e.key === "T") {
        ;(e as KeyboardEvent & Record<string, boolean>)[handledKey] = true
        e.preventDefault()
        setShowToc((prev) => !prev)
        return
      }

      if (code === "KeyS" || e.key === "s" || e.key === "S") {
        ;(e as KeyboardEvent & Record<string, boolean>)[handledKey] = true
        e.preventDefault()
        setShowSettings((prev) => !prev)
        return
      }

      if (code === "KeyB" || e.key === "b" || e.key === "B") {
        ;(e as KeyboardEvent & Record<string, boolean>)[handledKey] = true
        e.preventDefault()
        setShowBookmarks((prev) => !prev)
        return
      }

      // Keep navigation disabled while side panels are open.
      if (showSettings || showToc || showBookmarks) {
        return
      }

      if (e.key === "ArrowLeft" || e.key === "PageUp") {
        ;(e as KeyboardEvent & Record<string, boolean>)[handledKey] = true
        e.preventDefault()
        onPrev()
      } else if (
        e.key === "ArrowRight" ||
        e.key === "PageDown" ||
        e.key === " " ||
        e.key === "Spacebar"
      ) {
        ;(e as KeyboardEvent & Record<string, boolean>)[handledKey] = true
        e.preventDefault()
        onNext()
      }
    }

    const targets = new Set<Document>([document])

    const contents = viewRef.current?.renderer?.getContents?.() ?? []
    for (const content of contents) {
      targets.add(content.doc)
    }

    for (const target of targets) {
      target.addEventListener("keydown", handleKeyDown as EventListener, { capture: true })
    }

    const handleReaderLoad = () => {
      const nextContents = viewRef.current?.renderer?.getContents?.() ?? []
      for (const content of nextContents) {
        content.doc.addEventListener("keydown", handleKeyDown as EventListener, { capture: true })
      }
      viewRef.current?.renderer?.focusView?.()
    }

    const handleReaderPointerDown = () => {
      viewRef.current?.renderer?.focusView?.()
    }

    viewRef.current?.addEventListener("load", handleReaderLoad as EventListener)
    viewRef.current?.addEventListener("pointerdown", handleReaderPointerDown)

    return () => {
      for (const target of targets) {
        target.removeEventListener("keydown", handleKeyDown as EventListener, { capture: true })
      }
      const nextContents = viewRef.current?.renderer?.getContents?.() ?? []
      for (const content of nextContents) {
        content.doc.removeEventListener("keydown", handleKeyDown as EventListener, { capture: true })
      }
      viewRef.current?.removeEventListener("load", handleReaderLoad as EventListener)
      viewRef.current?.removeEventListener("pointerdown", handleReaderPointerDown)
    }
  }, [
    viewRef,
    readerReady,
    showSettings,
    showToc,
    showBookmarks,
    setShowSettings,
    setShowToc,
    setShowBookmarks,
    onPrev,
    onNext,
    onClose,
  ])
}
