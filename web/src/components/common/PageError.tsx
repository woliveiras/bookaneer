import { Button } from "../ui"

interface PageErrorProps {
  message: string
  onBack?: () => void
  backLabel?: string
}

export function PageError({ message, onBack, backLabel = "Go Back" }: PageErrorProps) {
  return (
    <div className="text-center py-12">
      <p className="text-destructive mb-4">{message}</p>
      {onBack && <Button onClick={onBack}>{backLabel}</Button>}
    </div>
  )
}
