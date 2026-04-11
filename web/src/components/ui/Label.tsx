import type { LabelHTMLAttributes } from "react"
import { cn } from "../../lib/utils"

export interface LabelProps extends LabelHTMLAttributes<HTMLLabelElement> {}

export function Label({ className, ...props }: LabelProps) {
  return (
    // biome-ignore lint/a11y/noLabelWithoutControl: Generic label component, control is provided via htmlFor by consumers
    <label
      className={cn(
        "text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70",
        className,
      )}
      {...props}
    />
  )
}
