import { formatBytes } from "../../lib/format"
import type { ColumnConfig, ColumnSchema } from "../../lib/types"
import { Badge } from "../ui"

const FORMAT_COLORS: Record<string, string> = {
  epub: "bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200",
  pdf: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
  mobi: "bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200",
  azw3: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200",
  cbz: "bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200",
  cbr: "bg-violet-100 text-violet-800 dark:bg-violet-900 dark:text-violet-200",
}

const LANGUAGE_COLORS: Record<string, string> = {
  en: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
  pt: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  es: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
  fr: "bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200",
  de: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200",
}

const COLOR_MAPS: Record<string, Record<string, string>> = {
  format: FORMAT_COLORS,
  language: LANGUAGE_COLORS,
}

function resolveColorClass(col: ColumnSchema, value: string): string | undefined {
  if (!col.colorHint) return undefined
  if (col.colorHint.type === "static") return col.colorHint.value
  const map = COLOR_MAPS[col.colorHint.value]
  return map?.[value.toLowerCase()]
}

interface DynamicCellProps {
  column: ColumnSchema
  row: Record<string, unknown>
}

export function DynamicCell({ column, row }: DynamicCellProps) {
  const raw = row[column.key]
  const value = raw != null ? String(raw) : (column.fallback ?? "")

  switch (column.renderType) {
    case "badge": {
      if (!value || value === "-") {
        return <span className="text-xs text-muted-foreground">{value}</span>
      }
      const colorClass = resolveColorClass(column, value)
      return (
        <Badge
          variant="secondary"
          className={`text-xs ${column.uppercase ? "uppercase" : ""} ${colorClass ?? ""}`}
        >
          {value}
        </Badge>
      )
    }

    case "size":
      return (
        <span className="text-xs tabular-nums">
          {typeof raw === "number" && raw > 0 ? formatBytes(raw) : (column.fallback ?? "-")}
        </span>
      )

    case "number":
      return <span className="text-xs tabular-nums">{value}</span>

    case "peers": {
      const seeders = typeof row.seeders === "number" ? row.seeders : 0
      const leechers = typeof row.leechers === "number" ? row.leechers : 0
      return (
        <span className="text-xs tabular-nums">
          {seeders > 0 && <span className="text-green-600 dark:text-green-400">{seeders}</span>}
          {seeders > 0 && leechers > 0 && <span className="text-muted-foreground">/</span>}
          {leechers > 0 && <span className="text-red-500 dark:text-red-400">{leechers}</span>}
          {seeders === 0 && leechers === 0 && <span className="text-muted-foreground">-</span>}
        </span>
      )
    }

    case "indexer":
      return (
        <Badge variant="outline" className="text-xs">
          {value}
        </Badge>
      )

    default:
      return <span className="text-sm">{value}</span>
  }
}

interface DynamicColumnHeaderProps {
  config: ColumnConfig
}

export function DynamicColumnHeader({ config }: DynamicColumnHeaderProps) {
  return (
    <div
      className="hidden sm:grid gap-2 px-4 py-1 text-xs font-medium text-muted-foreground uppercase tracking-wider"
      style={{ gridTemplateColumns: config.gridTemplate }}
    >
      {/* Title column header (first is always title) */}
      <span>Title</span>
      {config.columns.map((col) => (
        <span
          key={col.key}
          className={`${col.hideMobile ? "hidden md:block" : ""}`}
          style={{ textAlign: col.align }}
        >
          {col.label}
        </span>
      ))}
    </div>
  )
}
