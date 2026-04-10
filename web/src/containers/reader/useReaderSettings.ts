import { useCallback, useEffect, useState } from "react"
import {
  type FoliateView,
  loadSettings,
  type ReaderSettings,
  saveSettings,
  THEMES,
} from "../../components/reader/readerConfig"

/**
 * Manages reader display settings (font, theme, line height) and applies
 * them to the foliate-view renderer whenever they change.
 */
export function useReaderSettings(viewRef: React.RefObject<FoliateView | null>) {
  const [settings, setSettings] = useState<ReaderSettings>(loadSettings)

  const updateSettings = useCallback((updates: Partial<ReaderSettings>) => {
    setSettings((prev) => {
      const next = { ...prev, ...updates }
      saveSettings(next)
      return next
    })
  }, [])

  const applyStyles = useCallback(() => {
    const view = viewRef.current
    if (!view?.renderer?.setStyles) return
    const theme = THEMES[settings.theme]
    const css = `
      @import url('https://fonts.googleapis.com/css2?family=Literata:opsz,wght@7..72,400;7..72,700&display=swap');
      html { background: ${theme.bg} !important; color: ${theme.fg} !important; }
      body {
        font-family: ${settings.fontFamily} !important;
        font-size: ${settings.fontSize}% !important;
        line-height: ${settings.lineHeight} !important;
        background: ${theme.bg} !important;
        color: ${theme.fg} !important;
      }
      a { color: ${settings.theme === "dark" ? "#6ea8fe" : "#0d6efd"}; }
    `
    view.renderer.setStyles(css)
  }, [settings, viewRef])

  useEffect(() => {
    applyStyles()
  }, [applyStyles])

  return { settings, updateSettings, applyStyles }
}
