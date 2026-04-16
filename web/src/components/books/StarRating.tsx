import { Star } from "lucide-react"
import { useState } from "react"

interface StarRatingProps {
  value?: number // current rating 1-5, undefined = unrated
  onChange: (rating: number) => void
  disabled?: boolean
}

export function StarRating({ value, onChange, disabled }: StarRatingProps) {
  const [hovered, setHovered] = useState<number | null>(null)
  const [optimistic, setOptimistic] = useState<number | undefined>(undefined)

  // Sync optimistic state when the real value arrives from the server
  const committed = optimistic !== undefined ? optimistic : (value ?? 0)
  const display = hovered ?? committed

  const handleClick = (star: number) => {
    const next = committed === star ? 0 : star
    setOptimistic(next)
    onChange(next)
  }

  return (
    <fieldset className="flex gap-0.5 border-0 p-0 m-0" aria-label="Rating">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          disabled={disabled}
          aria-label={`Rate ${star} star${star > 1 ? "s" : ""}`}
          className="p-0.5 rounded transition-transform hover:scale-110 disabled:cursor-not-allowed"
          onClick={() => handleClick(star)}
          onMouseEnter={() => setHovered(star)}
          onMouseLeave={() => setHovered(null)}
        >
          <Star
            className={`w-5 h-5 transition-colors ${
              star <= display
                ? "fill-yellow-400 text-yellow-400"
                : "fill-transparent text-muted-foreground"
            }`}
          />
        </button>
      ))}
    </fieldset>
  )
}
