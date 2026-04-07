import type { Author } from "../../lib/api"
import { Card, CardContent, Badge } from "../ui"
import { cn } from "../../lib/utils"

interface AuthorCardProps {
  author: Author
  onClick?: () => void
  selected?: boolean
}

export function AuthorCard({ author, onClick, selected }: AuthorCardProps) {
  return (
    <Card
      className={cn(
        "cursor-pointer transition-colors hover:bg-accent/50",
        selected && "ring-2 ring-primary"
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
          {author.imageUrl ? (
            <img
              src={author.imageUrl}
              alt={`${author.name} portrait`}
              className="h-16 w-16 rounded-md object-cover"
              loading="lazy"
              width={64}
              height={64}
            />
          ) : (
            <div className="flex h-16 w-16 items-center justify-center rounded-md bg-muted">
              <span className="text-2xl" aria-hidden="true">
                👤
              </span>
            </div>
          )}
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold truncate">{author.name}</h3>
            <p className="text-sm text-muted-foreground truncate">{author.sortName}</p>
            <div className="mt-2 flex items-center gap-2">
              <Badge variant={author.monitored ? "success" : "secondary"}>
                {author.monitored ? "Monitored" : "Not Monitored"}
              </Badge>
              <Badge variant="outline">{author.status}</Badge>
            </div>
            {author.bookCount !== undefined && (
              <p className="mt-2 text-sm text-muted-foreground">
                {author.bookCount} book{author.bookCount !== 1 ? "s" : ""}
                {author.bookFileCount !== undefined && (
                  <> · {author.bookFileCount} file{author.bookFileCount !== 1 ? "s" : ""}</>
                )}
              </p>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
