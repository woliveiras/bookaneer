import { forwardRef, type ReactNode, useCallback, useEffect, useRef } from "react"
import { cn } from "../../lib/utils"

export interface DialogProps {
  open: boolean
  onClose: () => void
  children: ReactNode
  className?: string
}

export const Dialog = forwardRef<HTMLDialogElement, DialogProps>(
  ({ open, onClose, children, className }, ref) => {
    const innerRef = useRef<HTMLDialogElement>(null)
    const dialogRef = (ref as React.RefObject<HTMLDialogElement>) || innerRef

    // Sync open state with native dialog
    useEffect(() => {
      const el = dialogRef.current
      if (!el) return
      if (open && !el.open) {
        el.showModal()
      } else if (!open && el.open) {
        el.close()
      }
    }, [open, dialogRef])

    // Handle Escape and backdrop click
    const handleCancel = useCallback(
      (e: React.SyntheticEvent) => {
        e.preventDefault()
        onClose()
      },
      [onClose],
    )

    const handleBackdropClick = useCallback(
      (e: React.MouseEvent<HTMLDialogElement>) => {
        if (e.target === dialogRef.current) {
          onClose()
        }
      },
      [onClose, dialogRef],
    )

    return (
      <dialog
        ref={dialogRef}
        onCancel={handleCancel}
        onClick={handleBackdropClick}
        onKeyDown={(e) => {
          // Keyboard accessibility for backdrop - space/enter should not close
          // Escape is handled by onCancel event
          if (e.key === " " || e.key === "Enter") {
            e.stopPropagation()
          }
        }}
        className={cn(
          // Reset native dialog styles
          "fixed inset-0 m-auto p-0 border-0",
          // Sizing — large modal for search results
          "w-[95vw] max-w-4xl max-h-[90vh]",
          // Appearance
          "rounded-lg bg-background text-foreground shadow-xl",
          // Backdrop
          "backdrop:bg-black/60 backdrop:backdrop-blur-sm",
          className,
        )}
      >
        <div className="flex flex-col max-h-[90vh]">{children}</div>
      </dialog>
    )
  },
)
Dialog.displayName = "Dialog"

export function DialogHeader({
  children,
  className,
  onClose,
}: {
  children: ReactNode
  className?: string
  onClose?: () => void
}) {
  return (
    <div className={cn("flex items-center justify-between px-6 py-4 border-b shrink-0", className)}>
      <div className="flex-1 min-w-0">{children}</div>
      {onClose && (
        <button
          type="button"
          onClick={onClose}
          className="ml-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-hidden focus:ring-2 focus:ring-ring focus:ring-offset-2"
          aria-label="Close"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            <path d="M18 6 6 18" />
            <path d="m6 6 12 12" />
          </svg>
        </button>
      )}
    </div>
  )
}

export function DialogBody({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn("flex-1 overflow-y-auto px-6 py-4", className)}>{children}</div>
}
