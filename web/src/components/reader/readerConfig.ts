// Supported formats and their labels
export const FORMAT_LABELS: Record<string, string> = {
  epub: "EPUB",
  pdf: "PDF",
  mobi: "MOBI",
  azw3: "AZW3",
  fb2: "FB2",
  cbz: "CBZ",
}

// Theme presets
export const THEMES = {
  light: { bg: "#ffffff", fg: "#000000", name: "Light" },
  sepia: { bg: "#f4ecd8", fg: "#5b4636", name: "Sepia" },
  dark: { bg: "#1a1a1a", fg: "#e0e0e0", name: "Dark" },
} as const

export type ThemeKey = keyof typeof THEMES

// Font families
export const FONTS = [
  { value: "serif", label: "Serif" },
  { value: "sans-serif", label: "Sans Serif" },
  { value: "Georgia, serif", label: "Georgia" },
  { value: "'Literata', serif", label: "Literata" },
] as const

// Reader settings stored in localStorage
export interface ReaderSettings {
  theme: ThemeKey
  fontSize: number
  fontFamily: string
  lineHeight: number
}

export const DEFAULT_SETTINGS: ReaderSettings = {
  theme: "light",
  fontSize: 100,
  fontFamily: "serif",
  lineHeight: 1.5,
}

export function loadSettings(): ReaderSettings {
  try {
    const stored = localStorage.getItem("bookaneer_reader_settings")
    if (stored) {
      return { ...DEFAULT_SETTINGS, ...JSON.parse(stored) }
    }
  } catch {
    // Ignore parse errors
  }
  return DEFAULT_SETTINGS
}

export function saveSettings(settings: ReaderSettings): void {
  localStorage.setItem("bookaneer_reader_settings", JSON.stringify(settings))
}

export interface TocItem {
  label: string
  href: string
  subitems?: TocItem[]
}

export interface FoliateView extends HTMLElement {
  open: (source: Blob | string) => Promise<void>
  goTo: (target: string | number | { fraction: number }) => Promise<void>
  goToFraction: (fraction: number) => Promise<void>
  prev: () => Promise<void>
  next: () => Promise<void>
  renderer?: {
    setStyles?: (css: string | [string, string]) => void
    setAttribute?: (name: string, value: string) => void
    focusView?: () => void
    getContents?: () => Array<{
      index: number
      doc: Document
    }>
  }
  book?: {
    toc?: TocItem[]
    metadata?: {
      title?: string
      author?: string | string[]
    }
  }
}

export interface RelocateDetail {
  cfi?: string
  fraction?: number
  location?: {
    current: number
    total: number
  }
  tocItem?: {
    label?: string
  }
}
