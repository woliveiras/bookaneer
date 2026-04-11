import { beforeEach, describe, expect, it } from "vitest"
import {
  DEFAULT_SETTINGS,
  FONTS,
  FORMAT_LABELS,
  loadSettings,
  saveSettings,
  THEMES,
} from "./readerConfig"

beforeEach(() => {
  localStorage.clear()
})

describe("loadSettings", () => {
  it("returns defaults when nothing stored", () => {
    const settings = loadSettings()
    expect(settings).toEqual(DEFAULT_SETTINGS)
  })

  it("returns stored settings merged with defaults", () => {
    localStorage.setItem(
      "bookaneer_reader_settings",
      JSON.stringify({ theme: "dark", fontSize: 120 }),
    )
    const settings = loadSettings()
    expect(settings.theme).toBe("dark")
    expect(settings.fontSize).toBe(120)
    expect(settings.fontFamily).toBe("serif") // default
    expect(settings.lineHeight).toBe(1.5) // default
  })

  it("handles invalid JSON gracefully", () => {
    localStorage.setItem("bookaneer_reader_settings", "corrupted")
    const settings = loadSettings()
    expect(settings).toEqual(DEFAULT_SETTINGS)
  })
})

describe("saveSettings", () => {
  it("persists settings to localStorage", () => {
    const custom = { ...DEFAULT_SETTINGS, theme: "sepia" as const, fontSize: 110 }
    saveSettings(custom)
    const raw = localStorage.getItem("bookaneer_reader_settings")
    const stored = JSON.parse(raw ?? "{}")
    expect(stored.theme).toBe("sepia")
    expect(stored.fontSize).toBe(110)
  })
})

describe("constants", () => {
  it("THEMES has three entries", () => {
    expect(Object.keys(THEMES)).toEqual(["light", "sepia", "dark"])
  })

  it("FONTS has serif options", () => {
    expect(FONTS.length).toBeGreaterThanOrEqual(2)
    expect(FONTS[0].value).toBe("serif")
  })

  it("FORMAT_LABELS covers common ebook formats", () => {
    expect(FORMAT_LABELS.epub).toBe("EPUB")
    expect(FORMAT_LABELS.pdf).toBe("PDF")
    expect(FORMAT_LABELS.mobi).toBe("MOBI")
  })
})
