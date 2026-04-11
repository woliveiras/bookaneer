import { useCallback, useState } from "react"
import { User, Library } from "lucide-react"
import { Badge, Button, Card, CardContent, Input } from "../../components/ui"
import { useMetadataSearchAuthors, useMetadataSearchBooks } from "../../hooks/useMetadata"
import type { MetadataAuthorResult, MetadataBookResult } from "../../lib/api"

interface MetadataSearchProps {
  onSelectAuthor?: (author: MetadataAuthorResult) => void
  onSelectBook?: (book: MetadataBookResult) => void
}

export function MetadataSearch({ onSelectAuthor, onSelectBook }: MetadataSearchProps) {
  const [query, setQuery] = useState("")
  const [searchType, setSearchType] = useState<"authors" | "books">("books")
  const [debouncedQuery, setDebouncedQuery] = useState("")

  const handleSearch = useCallback(() => {
    setDebouncedQuery(query)
  }, [query])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        handleSearch()
      }
    },
    [handleSearch],
  )

  const authorSearch = useMetadataSearchAuthors(debouncedQuery, searchType === "authors")
  const bookSearch = useMetadataSearchBooks(debouncedQuery, searchType === "books")

  const isLoading = searchType === "authors" ? authorSearch.isLoading : bookSearch.isLoading
  const results = searchType === "authors" ? authorSearch.data?.results : bookSearch.data?.results

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <div className="flex-1">
          <Input
            type="search"
            placeholder={
              searchType === "authors"
                ? "Search authors by name..."
                : "Search books by title, author, or ISBN..."
            }
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            aria-label="Search query"
          />
        </div>
        <select
          className="rounded-md border border-input bg-background px-3 py-2 text-sm"
          value={searchType}
          onChange={(e) => setSearchType(e.target.value as "authors" | "books")}
          aria-label="Search type"
        >
          <option value="books">Books</option>
          <option value="authors">Authors</option>
        </select>
        <Button onClick={handleSearch} disabled={query.length < 2}>
          Search
        </Button>
      </div>

      {isLoading && (
        <div className="flex justify-center py-8">
          <div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
        </div>
      )}

      {searchType === "authors" && results && (
        <AuthorResults results={results as MetadataAuthorResult[]} onSelect={onSelectAuthor} />
      )}

      {searchType === "books" && results && (
        <BookResults results={results as MetadataBookResult[]} onSelect={onSelectBook} />
      )}

      {!isLoading && debouncedQuery && results?.length === 0 && (
        <p className="text-center text-muted-foreground py-8">
          No {searchType} found for "{debouncedQuery}"
        </p>
      )}
    </div>
  )
}

interface AuthorResultsProps {
  results: MetadataAuthorResult[]
  onSelect?: (author: MetadataAuthorResult) => void
}

function AuthorResults({ results, onSelect }: AuthorResultsProps) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {results.map((author) => (
        <Card
          key={`${author.provider}-${author.foreignId}`}
          className={onSelect ? "cursor-pointer hover:border-primary transition-colors" : ""}
          onClick={() => onSelect?.(author)}
        >
          <CardContent className="p-4 flex gap-4">
            {author.photoUrl ? (
              <img
                src={author.photoUrl}
                alt={author.name}
                className="w-16 h-20 object-cover rounded"
                loading="lazy"
              />
            ) : (
              <div className="w-16 h-20 bg-muted rounded flex items-center justify-center">
                <User className="w-6 h-6 text-muted-foreground" />
              </div>
            )}
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold truncate">{author.name}</h3>
              {(author.birthYear || author.deathYear) && (
                <p className="text-sm text-muted-foreground">
                  {author.birthYear || "?"} - {author.deathYear || "present"}
                </p>
              )}
              {author.worksCount !== undefined && author.worksCount > 0 && (
                <p className="text-sm text-muted-foreground">{author.worksCount} works</p>
              )}
              <Badge variant="outline" className="mt-2">
                {author.provider}
              </Badge>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

interface BookResultsProps {
  results: MetadataBookResult[]
  onSelect?: (book: MetadataBookResult) => void
}

function BookResults({ results, onSelect }: BookResultsProps) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {results.map((book) => (
        <Card
          key={`${book.provider}-${book.foreignId}`}
          className={onSelect ? "cursor-pointer hover:border-primary transition-colors" : ""}
          onClick={() => onSelect?.(book)}
        >
          <CardContent className="p-4 flex gap-4">
            {book.coverUrl ? (
              <img
                src={book.coverUrl}
                alt={book.title}
                className="w-16 h-24 object-cover rounded"
                loading="lazy"
              />
            ) : (
              <div className="w-16 h-24 bg-muted rounded flex items-center justify-center">
                <Library className="w-6 h-6 text-muted-foreground" />
              </div>
            )}
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold line-clamp-2">{book.title}</h3>
              {book.authors && book.authors.length > 0 && (
                <p className="text-sm text-muted-foreground truncate">{book.authors.join(", ")}</p>
              )}
              {book.publishedYear && (
                <p className="text-sm text-muted-foreground">{book.publishedYear}</p>
              )}
              <div className="flex gap-2 mt-2 flex-wrap">
                <Badge variant="outline">{book.provider}</Badge>
                {book.isbn13 && (
                  <Badge variant="secondary" className="font-mono text-xs">
                    {book.isbn13}
                  </Badge>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
