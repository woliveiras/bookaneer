import { useState } from "react"
import { useSearch, type SearchParams } from "../../hooks/useIndexers"
import type { SearchResult } from "../../lib/api"
import { Button, Input, Card, CardContent, Badge } from "../ui"

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B"
  const k = 1024
  const sizes = ["B", "KB", "MB", "GB"]
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / k ** i).toFixed(1))} ${sizes[i]}`
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" })
}

function ResultCard({ result }: { result: SearchResult }) {
  return (
    <Card>
      <CardContent className="py-4">
        <div className="flex justify-between items-start gap-4">
          <div className="flex-1 min-w-0">
            <h4 className="font-medium truncate">{result.title}</h4>
            {result.description && (
              <p className="text-sm text-muted-foreground line-clamp-2 mt-1">{result.description}</p>
            )}
            <div className="flex flex-wrap gap-2 mt-2">
              <Badge variant="outline">{result.indexerName}</Badge>
              <Badge variant="secondary">{formatBytes(result.size)}</Badge>
              {result.category && <Badge variant="secondary">{result.category}</Badge>}
              {result.seeders !== undefined && result.seeders > 0 && (
                <Badge variant="default" className="bg-green-600">
                  {result.seeders} seeders
                </Badge>
              )}
              {result.leechers !== undefined && result.leechers > 0 && (
                <Badge variant="secondary">{result.leechers} leechers</Badge>
              )}
              {result.grabs !== undefined && result.grabs > 0 && (
                <Badge variant="outline">{result.grabs} grabs</Badge>
              )}
            </div>
          </div>
          <div className="flex flex-col items-end gap-2">
            <span className="text-xs text-muted-foreground">{formatDate(result.pubDate)}</span>
            <div className="flex gap-2">
              {result.infoUrl && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => window.open(result.infoUrl, "_blank")}
                >
                  Info
                </Button>
              )}
              <Button
                variant="default"
                size="sm"
                onClick={() => window.open(result.downloadUrl, "_blank")}
              >
                Download
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export function InteractiveSearch() {
  const [searchParams, setSearchParams] = useState<SearchParams>({})
  const [submittedParams, setSubmittedParams] = useState<SearchParams>({})
  const { data, isLoading, error } = useSearch(submittedParams)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSubmittedParams(searchParams)
  }

  const hasQuery = !!(searchParams.q || searchParams.author || searchParams.title || searchParams.isbn)

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            placeholder="Search query..."
            value={searchParams.q ?? ""}
            onChange={(e) => setSearchParams({ ...searchParams, q: e.target.value })}
          />
          <Input
            placeholder="Author name..."
            value={searchParams.author ?? ""}
            onChange={(e) => setSearchParams({ ...searchParams, author: e.target.value })}
          />
          <Input
            placeholder="Book title..."
            value={searchParams.title ?? ""}
            onChange={(e) => setSearchParams({ ...searchParams, title: e.target.value })}
          />
          <Input
            placeholder="ISBN..."
            value={searchParams.isbn ?? ""}
            onChange={(e) => setSearchParams({ ...searchParams, isbn: e.target.value })}
          />
        </div>
        <div className="flex gap-2">
          <Button type="submit" disabled={!hasQuery || isLoading}>
            {isLoading ? "Searching..." : "Search"}
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              setSearchParams({})
              setSubmittedParams({})
            }}
          >
            Clear
          </Button>
        </div>
      </form>

      {error && (
        <div className="text-destructive">Error: {error.message}</div>
      )}

      {data && (
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">
            {data.total} {data.total === 1 ? "result" : "results"} found
          </h3>
          {data.results.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                No results found. Try adjusting your search terms.
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4">
              {data.results.map((result) => (
                <ResultCard key={result.guid} result={result} />
              ))}
            </div>
          )}
        </div>
      )}

      {!data && !isLoading && (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            Enter a search query to find ebooks across your configured indexers.
          </CardContent>
        </Card>
      )}
    </div>
  )
}
