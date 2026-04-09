import type { TocItem } from "./readerConfig"

interface TocListProps {
  items: TocItem[]
  onSelect: (href: string) => void
  theme: { bg: string; fg: string }
  depth?: number
}

export function TocList({ items, onSelect, theme, depth = 0 }: TocListProps) {
  return (
    <ul className={depth > 0 ? "ml-4 mt-1" : ""}>
      {items.map((item, index) => (
        <li key={`${item.href}-${index}`}>
          <button
            type="button"
            onClick={() => onSelect(item.href)}
            className="w-full text-left py-1.5 px-2 rounded hover:bg-black/10 dark:hover:bg-white/10 text-sm"
            style={{ color: theme.fg }}
          >
            {item.label}
          </button>
          {item.subitems && item.subitems.length > 0 && (
            <TocList items={item.subitems} onSelect={onSelect} theme={theme} depth={depth + 1} />
          )}
        </li>
      ))}
    </ul>
  )
}
