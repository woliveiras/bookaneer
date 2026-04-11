import { X } from "lucide-react"
import { FONTS, type ReaderSettings, THEMES, type ThemeKey } from "../../components/reader/readerConfig"
import { Button } from "../../components/ui"

interface ReaderSettingsPanelProps {
  settings: ReaderSettings
  onUpdateSettings: (updates: Partial<ReaderSettings>) => void
  onClose: () => void
}

export function ReaderSettingsPanel({
  settings,
  onUpdateSettings,
  onClose,
}: ReaderSettingsPanelProps) {
  const theme = THEMES[settings.theme]
  const borderColor = settings.theme === "dark" ? "#333" : settings.theme === "sepia" ? "#d4c9b0" : "#e5e5e5"

  return (
    <div
      className="absolute right-0 top-0 bottom-0 w-80 overflow-y-auto border-l shadow-lg"
      style={{ background: theme.bg, borderColor }}
    >
      <div className="p-4">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-bold text-lg">Settings</h2>
          <Button variant="ghost" size="sm" onClick={onClose} aria-label="Close settings" style={{ color: theme.fg }}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Theme */}
        <div className="mb-6">
          <label className="block text-sm font-medium mb-2">Theme</label>
          <div className="flex gap-2">
            {(Object.keys(THEMES) as ThemeKey[]).map((key) => (
              <button
                key={key}
                type="button"
                onClick={() => onUpdateSettings({ theme: key })}
                className={`flex-1 py-2 px-3 rounded border text-sm ${
                  settings.theme === key ? "ring-2 ring-blue-500" : ""
                }`}
                style={{
                  background: THEMES[key].bg,
                  color: THEMES[key].fg,
                  borderColor: settings.theme === "dark" ? "#555" : "#ccc",
                }}
              >
                {THEMES[key].name}
              </button>
            ))}
          </div>
        </div>

        {/* Font Size */}
        <div className="mb-6">
          <label className="block text-sm font-medium mb-2">
            Font Size: {settings.fontSize}%
          </label>
          <input
            type="range"
            min="75"
            max="200"
            step="5"
            value={settings.fontSize}
            onChange={(e) => onUpdateSettings({ fontSize: Number(e.target.value) })}
            className="w-full"
          />
          <div className="flex justify-between text-xs" style={{ opacity: 0.7 }}>
            <span>75%</span>
            <span>200%</span>
          </div>
        </div>

        {/* Font Family */}
        <div className="mb-6">
          <label className="block text-sm font-medium mb-2">Font</label>
          <select
            value={settings.fontFamily}
            onChange={(e) => onUpdateSettings({ fontFamily: e.target.value })}
            className="w-full p-2 rounded border"
            style={{
              background: theme.bg,
              color: theme.fg,
              borderColor: settings.theme === "dark" ? "#555" : "#ccc",
            }}
          >
            {FONTS.map((font) => (
              <option key={font.value} value={font.value}>
                {font.label}
              </option>
            ))}
          </select>
        </div>

        {/* Line Height */}
        <div className="mb-6">
          <label className="block text-sm font-medium mb-2">
            Line Height: {settings.lineHeight.toFixed(1)}
          </label>
          <input
            type="range"
            min="1"
            max="2.5"
            step="0.1"
            value={settings.lineHeight}
            onChange={(e) => onUpdateSettings({ lineHeight: Number(e.target.value) })}
            className="w-full"
          />
          <div className="flex justify-between text-xs" style={{ opacity: 0.7 }}>
            <span>1.0</span>
            <span>2.5</span>
          </div>
        </div>

        {/* Keyboard shortcuts */}
        <div className="mt-8 pt-4 border-t" style={{ borderColor }}>
          <h3 className="font-medium mb-2">Keyboard Shortcuts</h3>
          <dl className="text-sm space-y-1" style={{ opacity: 0.7 }}>
            <div className="flex justify-between">
              <dt>Previous page</dt>
              <dd>← / PageUp</dd>
            </div>
            <div className="flex justify-between">
              <dt>Next page</dt>
              <dd>→ / PageDown / Space</dd>
            </div>
            <div className="flex justify-between">
              <dt>Table of contents</dt>
              <dd>T</dd>
            </div>
            <div className="flex justify-between">
              <dt>Bookmarks</dt>
              <dd>B</dd>
            </div>
            <div className="flex justify-between">
              <dt>Settings</dt>
              <dd>S</dd>
            </div>
            <div className="flex justify-between">
              <dt>Close reader</dt>
              <dd>Esc</dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  )
}
