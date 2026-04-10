interface PageLoadingProps {
  message?: string
}

export function PageLoading({ message = "Loading..." }: PageLoadingProps) {
  return (
    <div className="flex items-center justify-center py-12">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
        <p className="text-muted-foreground">{message}</p>
      </div>
    </div>
  )
}
