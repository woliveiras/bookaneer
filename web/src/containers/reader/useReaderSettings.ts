import { useCallback, useEffect } from "react"
import { useShallow } from "zustand/react/shallow"
import {
  type FoliateView,
  THEMES,
} from "../../components/reader/readerConfig"
import { useReaderSettingsStore } from "../../store/reader/reader-settings.store"

/**
 * Manages reader display settings (font, theme, line height) and applies
 * them to the foliate-view renderer whenever they change.
 */
export function useReaderSettings(viewRef: React.RefObject<FoliateView | null>) {
  const settings = useReaderSettingsStore(
    useShallow((s) => ({
      theme: s.theme,
      fontSize: s.fontSize,
      fontFamily: s.fontFamily,
      lineHeight: s.lineHeight,
    })),
  )
  const updateSettings = useReaderSettingsStore((s) => s.updateSettings)

  const applyStyles = useCallback(() => {
    const view = viewRef.current
    if (!view?.renderer?.setStyles) return
    const theme = THEMES[settings.theme]
    const isDark = settings.theme === "dark"
    const borderColor = isDark ? "#404040" : "#e0e0e0"
    const codeBg = isDark ? "#2d2d2d" : "#f5f5f5"
    const codeFg = isDark ? "#d4d4d4" : "#1e1e1e"
    const linkColor = isDark ? "#6ea8fe" : "#0d6efd"

    // "before" styles: low-specificity defaults that the book's own CSS can override
    const beforeCss = `
      table {
        display: table;
        border-collapse: collapse;
        width: 100%;
        margin: 1em 0;
        page-break-inside: avoid;
        break-inside: avoid;
      }
      thead { display: table-header-group; }
      tbody { display: table-row-group; }
      tr { display: table-row; }
      th, td {
        display: table-cell;
        padding: 0.4em 0.6em;
        border: 1px solid ${borderColor};
        vertical-align: top;
        text-align: left;
      }
      th {
        font-weight: bold;
        background: ${isDark ? "#333" : "#f0f0f0"};
      }
      caption {
        caption-side: top;
        font-weight: bold;
        padding: 0.5em 0;
      }
    `

    // "after" styles: user preferences that override the book's CSS
    const afterCss = `
      @import url('https://fonts.googleapis.com/css2?family=Literata:opsz,wght@7..72,400;7..72,700&display=swap');
      html {
        background: ${theme.bg} !important;
        color: ${theme.fg} !important;
      }
      body {
        font-family: ${settings.fontFamily};
        font-size: ${settings.fontSize}% !important;
        line-height: ${settings.lineHeight} !important;
        background: ${theme.bg} !important;
        color: ${theme.fg} !important;
      }
      a { color: ${linkColor}; }
      pre, code, kbd, samp, .programlisting, .literal, .code {
        font-family: "SFMono-Regular", "Menlo", "Consolas", "Liberation Mono", monospace !important;
        font-size: 0.9em;
      }
      pre, .programlisting {
        display: block !important;
        white-space: pre-wrap !important;
        word-wrap: break-word;
        overflow-wrap: break-word;
        padding: 0.75em 1em;
        margin: 1em 0;
        border-radius: 4px;
        background: ${codeBg} !important;
        color: ${codeFg} !important;
        border: 1px solid ${borderColor};
        line-height: 1.45 !important;
      }
      code {
        padding: 0.15em 0.3em;
        border-radius: 3px;
        background: ${isDark ? "#2d2d2d" : "#f0f0f0"};
      }
      pre code, .programlisting code {
        padding: 0;
        background: transparent;
        border: none;
      }
      blockquote {
        border-left: 3px solid ${isDark ? "#555" : "#ccc"};
        padding-left: 1em;
        margin-left: 0;
        font-style: italic;
      }
    `
    view.renderer.setStyles([beforeCss, afterCss])
  }, [settings, viewRef])

  useEffect(() => {
    applyStyles()
  }, [applyStyles])

  return { settings, updateSettings, applyStyles }
}
