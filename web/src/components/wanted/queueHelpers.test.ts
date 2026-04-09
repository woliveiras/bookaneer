import { beforeEach, describe, expect, it } from "vitest"
import type { ActiveCommand } from "../../lib/api"
import {
  getCommandDescription,
  getDismissedCommands,
  getStatusInfo,
  saveDismissedCommands,
} from "./queueHelpers"

beforeEach(() => {
  localStorage.clear()
})

describe("getDismissedCommands", () => {
  it("returns empty set when nothing stored", () => {
    const dismissed = getDismissedCommands()
    expect(dismissed.size).toBe(0)
  })

  it("returns stored command IDs", () => {
    localStorage.setItem("bookaneer-dismissed-commands", JSON.stringify(["cmd-1", "cmd-2"]))
    const dismissed = getDismissedCommands()
    expect(dismissed.size).toBe(2)
    expect(dismissed.has("cmd-1")).toBe(true)
  })

  it("handles invalid JSON gracefully", () => {
    localStorage.setItem("bookaneer-dismissed-commands", "not-json")
    const dismissed = getDismissedCommands()
    expect(dismissed.size).toBe(0)
  })
})

describe("saveDismissedCommands", () => {
  it("persists command IDs to localStorage", () => {
    saveDismissedCommands(new Set(["cmd-1", "cmd-3"]))
    const stored = JSON.parse(localStorage.getItem("bookaneer-dismissed-commands")!)
    expect(stored).toEqual(expect.arrayContaining(["cmd-1", "cmd-3"]))
  })
})

describe("getCommandDescription", () => {
  it("returns book info for AutomaticSearch", () => {
    const cmd: ActiveCommand = {
      id: "cmd-1",
      name: "AutomaticSearch",
      status: "running",
      priority: 1,
      trigger: "manual",
      queuedAt: "2025-01-01T00:00:00Z",
      payload: { bookTitle: "Dune", authorName: "Herbert" },
    }
    const desc = getCommandDescription(cmd)
    expect(desc.title).toBe("Dune")
    expect(desc.subtitle).toBe("by Herbert")
  })

  it("returns generic title for MissingBookSearch", () => {
    const cmd: ActiveCommand = {
      id: "cmd-2",
      name: "MissingBookSearch",
      status: "running",
      priority: 1,
      trigger: "manual",
      queuedAt: "2025-01-01T00:00:00Z",
      payload: {},
    }
    const desc = getCommandDescription(cmd)
    expect(desc.title).toBe("All Missing Books")
  })

  it("falls back to command name for unknown commands", () => {
    const cmd: ActiveCommand = {
      id: "cmd-3",
      name: "CustomCommand",
      status: "running",
      priority: 1,
      trigger: "manual",
      queuedAt: "2025-01-01T00:00:00Z",
      payload: {},
    }
    const desc = getCommandDescription(cmd)
    expect(desc.title).toBe("CustomCommand")
  })
})

describe("getStatusInfo", () => {
  it("returns searching state for running commands", () => {
    const info = getStatusInfo("running", false)
    expect(info.label).toBe("Searching...")
    expect(info.spinning).toBe(true)
  })

  it("returns not found for failed commands", () => {
    const info = getStatusInfo("failed", false)
    expect(info.label).toBe("Not Found")
    expect(info.spinning).toBe(false)
  })

  it("returns not found when hasError is true", () => {
    const info = getStatusInfo("completed", true)
    expect(info.label).toBe("Not Found")
  })

  it("returns found for completed commands without errors", () => {
    const info = getStatusInfo("completed", false)
    expect(info.label).toBe("Found")
    expect(info.spinning).toBe(false)
  })
})
