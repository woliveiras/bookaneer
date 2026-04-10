import { useEffect } from "react"

interface UseReaderKeyboardProps {
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
      if (showSettings || showToc || showBookmarks) {
        if (e.key === "Escape") {
          setShowSettings(false)
          setShowToc(false)
          setShowBookmarks(false)
        }
        return
      }

      if (e.key === "ArrowLeft" || e.key === "PageUp") {
        e.preventDefault()
        onPrev()
      } else if (e.key === "ArrowRight" || e.key === "PageDown" || e.key === " ") {
        e.preventDefault()
        onNext()
      } else if (e.key === "Escape") {
        e.preventDefault()
        onClose()
      } else if (e.key === "t" || e.key === "T") {
        e.preventDefault()
        setShowToc((prev) => !prev)
      } else if (e.key === "s" || e.key === "S") {
        e.preventDefault()
        setShowSettings((prev) => !prev)
      } else if (e.key === "b" || e.key === "B") {
        e.preventDefault()
        setShowBookmarks((prev) => !prev)
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [
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
