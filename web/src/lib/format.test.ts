import { describe, expect, it } from "vitest"
import { formatBytes } from "./format"

describe("formatBytes", () => {
  it("returns '0 B' for 0 bytes", () => {
    expect(formatBytes(0)).toBe("0 B")
  })

  it("formats bytes", () => {
    expect(formatBytes(500)).toBe("500 B")
  })

  it("formats kilobytes", () => {
    expect(formatBytes(1024)).toBe("1 KB")
    expect(formatBytes(1536)).toBe("1.5 KB")
  })

  it("formats megabytes", () => {
    expect(formatBytes(1048576)).toBe("1 MB")
    expect(formatBytes(5242880)).toBe("5 MB")
  })

  it("formats gigabytes", () => {
    expect(formatBytes(1073741824)).toBe("1 GB")
  })
})
