import { Link } from "@tanstack/react-router"
import { BookOpen } from "lucide-react"
import type { Book } from "../../lib/api"
import { cn } from "../../lib/utils"
import { Badge, Button, Card, CardContent } from "../ui"

interface BookCardProps {
  book: Book
  onClick?: () => void
  selected?: boolean
  actions?: React.ReactNode
}

export function BookCard({ book, onClick, selected, actions }: BookCardProps) {
  return (
    <Card
      className={cn(
        "flex flex-col overflow-hidden transition-colors h-full",
        onClick && "cursor-pointer hover:bg-accent/50",
        selected && "ring-2 ring-primary",
      )}
      onClick={onClick}
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
      onKeyDown={
        onClick
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault()
                onClick()
              }
            }
          : undefined
      }
    >
      {/* Cover */}
      <div className="w-full h-40 bg-muted overflow-hidden shrink-0">
        {book.imageUrl ? (
          <img
            src={book.imageUrl}
            alt={`${book.title} cover`}
            className="w-full h-full object-cover"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <BookOpen className="w-10 h-10 text-muted-foreground" aria-hidden="true" />
          </div>
        )}
      </div>

      <CardContent className="flex flex-col flex-1 gap-3 p-4">
        {/* Book info */}
        <div className="flex-1 min-w-0 space-y-1">
          <h3 className="font-semibold line-clamp-2 leading-snug">{book.title}</h3>
          {book.authorName && <p className="text-sm text-muted-foreground">by {book.authorName}</p>}
          <div className="flex flex-wrap items-center gap-2 mt-2">
            {book.hasFile ? (
              <Badge variant="default">Has File</Badge>
            ) : (
              <Badge variant="outline">Missing</Badge>
            )}
            {book.fileFormat && (
              <Badge variant="secondary" className="uppercase">
                {book.fileFormat}
              </Badge>
            )}
          </div>
          {book.releaseDate && (
            <p className="text-xs text-muted-foreground">
              Released: {new Date(book.releaseDate).toLocaleDateString()}
            </p>
          )}
          {book.isbn13 && (
            <p className="text-xs text-muted-foreground font-mono">ISBN: {book.isbn13}</p>
          )}
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          {actions ??
            (book.hasFile ? (
              <Link
                to="/read/$bookId"
                params={{ bookId: String(book.id) }}
                onClick={(e) => e.stopPropagation()}
              >
                <Button
                  variant="default"
                  size="sm"
                  className="inline-flex items-center gap-2"
                  aria-label={`Read ${book.title}`}
                >
                  <BookOpen className="h-4 w-4" aria-hidden="true" />
                  Read
                </Button>
              </Link>
            ) : null)}
        </div>
      </CardContent>
    </Card>
  )
}
