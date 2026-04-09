import { Link } from "@tanstack/react-router"
import type { Book } from "../../lib/api"
import { cn } from "../../lib/utils"
import { Badge, Button, Card, CardContent } from "../ui"

interface BookCardProps {
  book: Book
  onClick?: () => void
  selected?: boolean
}

export function BookCard({ book, onClick, selected }: BookCardProps) {
  return (
    <Card
      className={cn(
        "cursor-pointer transition-colors hover:bg-accent/50",
        selected && "ring-2 ring-primary",
      )}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault()
          onClick?.()
        }
      }}
    >
      <CardContent className="p-4">
        <div className="flex items-start gap-4">
          {book.imageUrl ? (
            <img
              src={book.imageUrl}
              alt={`${book.title} cover`}
              className="h-24 w-16 rounded object-cover"
              loading="lazy"
              width={64}
              height={96}
            />
          ) : (
            <div className="flex h-24 w-16 items-center justify-center rounded bg-muted">
              <span className="text-2xl" aria-hidden="true">
                📖
              </span>
            </div>
          )}
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold truncate">{book.title}</h3>
            {book.authorName && (
              <p className="text-sm text-muted-foreground truncate">by {book.authorName}</p>
            )}
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <Badge variant={book.monitored ? "success" : "secondary"}>
                {book.monitored ? "Monitored" : "Not Monitored"}
              </Badge>
              {book.hasFile ? (
                <Badge variant="default">Has File</Badge>
              ) : (
                <Badge variant="outline">Missing</Badge>
              )}
            </div>
            {book.releaseDate && (
              <p className="mt-2 text-sm text-muted-foreground">
                Released: {new Date(book.releaseDate).toLocaleDateString()}
              </p>
            )}
            {book.isbn13 && (
              <p className="text-xs text-muted-foreground font-mono">ISBN: {book.isbn13}</p>
            )}
            {book.hasFile && (
              <Link
                to="/read/$bookId"
                params={{ bookId: String(book.id) }}
                onClick={(e) => e.stopPropagation()}
              >
                <Button
                  variant="default"
                  size="sm"
                  className="mt-2"
                  aria-label={`Read ${book.title}`}
                >
                  📖 Read
                </Button>
              </Link>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
