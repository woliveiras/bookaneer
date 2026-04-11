import { useEffect } from "react"
import type { FoliateView } from "../../components/reader/readerConfig"

interface UseReaderKeyboardProps {
  viewRef: React.RefObject<FoliateView | null>
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
    const handleKeyDown = (e: KeyboardEvent) => {
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

      if (e.key === "t" || e.key === "T") {
        e.preventDefault()
        setShowToc((prev) => !prev)
        return
      }

      if (e.key === "s" || e.key === "S") {
        e.preventDefault()
        setShowSettings((prev) => !prev)
        return
      }

      if (e.key === "b" || e.key === "B") {
        e.preventDefault()
        setShowBookmarks((prev) => !prev)
        return
      }

      // Keep navigation disabled while side panels are open.
      if (showSettings || showToc || showBookmarks) {
        return
      }

      if (e.key === "ArrowLeft" || e.key === "PageUp") {
        e.preventDefault()
        onPrev()
      } else if (
        e.key === "ArrowRight" ||
        e.key === "PageDown" ||
        e.key === " " ||
        e.key === "Spacebar"
      ) {
        e.preventDefault()
        onNext()
      }
    }

    const targets = new Set<Window | Document>([window, document])

    const contents = viewRef.current?.renderer?.getContents?.() ?? []
    for (const content of contents) {
      targets.add(content.doc)
      if (content.doc.defaultView) {
        targets.add(content.doc.defaultView)
      }
    }

    for (const target of targets) {
      target.addEventListener("keydown", handleKeyDown as EventListener, { capture: true })
    }

    const handleReaderLoad = () => {
      const nextContents = viewRef.current?.renderer?.getContents?.() ?? []
      for (const content of nextContents) {
        content.doc.addEventListener("keydown", handleKeyDown as EventListener, { capture: true })
        content.doc.defaultView?.addEventListener("keydown", handleKeyDown as EventListener, {
          capture: true,
        })
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
        content.doc.defaultView?.removeEventListener(
          "keydown",
          handleKeyDown as EventListener,
          { capture: true },
        )
      }
      viewRef.current?.removeEventListener("load", handleReaderLoad as EventListener)
      viewRef.current?.removeEventListener("pointerdown", handleReaderPointerDown)
    }
  }, [
    viewRef,
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
