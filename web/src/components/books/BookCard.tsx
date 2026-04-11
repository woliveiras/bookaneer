import { BookOpen } from "lucide-react"
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
        "h-full cursor-pointer transition-colors hover:bg-accent/50",
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
      <CardContent className="flex h-full flex-col p-4">
        <div className="flex flex-1 items-start gap-4">
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
              <BookOpen className="h-6 w-6 text-muted-foreground" aria-hidden="true" />
            </div>
          )}
          <div className="flex min-w-0 flex-1 flex-col">
            <div>
              <h3 className="line-clamp-2 min-h-[3.5rem] font-semibold leading-7">{book.title}</h3>
            </div>
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
              {book.fileFormat && (
                <Badge variant="secondary" className="uppercase">
                  {book.fileFormat}
                </Badge>
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
          </div>
        </div>
        <div className="mt-4 min-h-9">
          {book.hasFile ? (
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
          ) : null}
        </div>
      </CardContent>
    </Card>
  )
}
