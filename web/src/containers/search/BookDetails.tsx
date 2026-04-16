import { useNavigate } from "@tanstack/react-router"
import { BookHeader } from "../../components/search/BookHeader"
import { useMetadataBook } from "../../hooks/useMetadata"
import { useRootFolders } from "../../hooks/useRootFolders"
import type { MetadataBookResult } from "../../lib/api"

interface BookDetailsProps {
  book: MetadataBookResult
  autoSearch?: boolean
  existingBookId?: number
}

export function BookDetails({ book, existingBookId }: BookDetailsProps) {
  const navigate = useNavigate()
  const { data: rootFolders } = useRootFolders()
  const { data: bookMetadata } = useMetadataBook(book.foreignId, book.provider, !!book.foreignId)

  const goToReleases = () => {
    void navigate({
      to: "/search/releases",
      search: {
        title: book.title,
        authors: book.authors?.join("|||"),
        provider: book.provider,
        foreignId: book.foreignId,
        publishedYear: book.publishedYear?.toString(),
        coverUrl: book.coverUrl,
        isbn13: book.isbn13,
        autoSearch: "true",
        bookId: existingBookId?.toString(),
      },
    })
  }

  return (
    <div className="space-y-6">
      <BookHeader
        book={book}
        bookMetadata={bookMetadata}
        addedToLibrary={false}
        addingToLibrary={false}
        addError={null}
        grabError={null}
        searchStarted={false}
        hasRootFolder={!!(rootFolders?.length)}
        onAddToLibrary={() => undefined}
        onStartSearch={goToReleases}
      />
    </div>
  )
}

