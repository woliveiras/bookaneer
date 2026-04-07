import { useState } from "react"
import { useAuthors } from "../../hooks/useAuthors"
import { AuthorCard } from "./AuthorCard"
import { Input, Button } from "../ui"

export function AuthorList() {
  const [search, setSearch] = useState("")
  const [debouncedSearch, setDebouncedSearch] = useState("")

  const { data, isLoading, error } = useAuthors({
    search: debouncedSearch || undefined,
    limit: 50,
  })

  // Simple debounce
  const handleSearch = (value: string) => {
    setSearch(value)
    const timeoutId = setTimeout(() => {
      setDebouncedSearch(value)
    }, 300)
    return () => clearTimeout(timeoutId)
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive bg-destructive/10 p-4" role="alert">
        <p className="text-destructive">Failed to load authors: {error.message}</p>
        <Button variant="outline" className="mt-2" onClick={() => window.location.reload()}>
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-4">
        <div className="flex-1 max-w-md">
          <label htmlFor="author-search" className="sr-only">
            Search authors
          </label>
          <Input
            id="author-search"
            type="search"
            placeholder="Search by name..."
            value={search}
            onChange={(e) => handleSearch(e.target.value)}
          />
        </div>
        <Button>Add Author</Button>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[...Array(6)].map((_, i) => (
            <div
              key={i}
              className="h-28 animate-pulse rounded-lg border bg-muted"
              aria-hidden="true"
            />
          ))}
        </div>
      ) : !data?.records?.length ? (
        <div className="rounded-lg border border-dashed p-8 text-center">
          <p className="text-muted-foreground">
            {debouncedSearch ? `No authors found for "${debouncedSearch}"` : "No authors yet"}
          </p>
          <Button variant="link" className="mt-2">
            Add your first author
          </Button>
        </div>
      ) : (
        <>
          <p className="text-sm text-muted-foreground">
            {data.totalRecords} author{data.totalRecords !== 1 ? "s" : ""}
          </p>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3" role="list">
            {data.records.map((author) => (
              <div key={author.id} role="listitem">
                <AuthorCard author={author} onClick={() => console.log("Selected:", author.id)} />
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
