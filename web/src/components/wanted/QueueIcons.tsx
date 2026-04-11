// Animated download icon for "downloading" status
export function DownloadingIcon() {
  return (
    <div className="relative flex items-center justify-center h-6 w-6">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-5 w-5 text-blue-500"
      >
        <title>Downloading</title>
        {/* Arrow pointing down with animation */}
        <path d="M12 3v12" className="animate-pulse" />
        <path d="m8 11 4 4 4-4" className="animate-bounce" />
        {/* Base line */}
        <path d="M8 21h8" />
        <path d="M12 17v4" />
      </svg>
      {/* Animated ring */}
      <div className="absolute inset-0 rounded-full border-2 border-blue-500/30 animate-ping" />
    </div>
  )
}

// Icon for "queued" status - waiting in line
export function QueuedIcon() {
  return (
    <div className="relative flex items-center justify-center h-6 w-6 rounded-full bg-amber-500">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="white"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-3.5 w-3.5"
      >
        <title>Queued</title>
        {/* Download arrow */}
        <path d="M12 5v8" />
        <path d="m8 10 4 4 4-4" />
        <path d="M5 19h14" />
      </svg>
      {/* Subtle pulse animation */}
      <div
        className="absolute inset-0 rounded-full bg-amber-400 animate-pulse opacity-50"
        style={{ animationDuration: "2s" }}
      />
    </div>
  )
}

// Animated search icon for "searching" status
export function SearchingIcon() {
  return (
    <div className="relative flex items-center justify-center h-6 w-6">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="h-5 w-5 text-blue-500 animate-pulse"
      >
        <title>Searching</title>
        {/* Magnifying glass */}
        <circle cx="11" cy="11" r="8" />
        <path d="m21 21-4.3-4.3" />
      </svg>
      {/* Animated ring */}
      <div
        className="absolute inset-0 rounded-full border-2 border-blue-500/30 animate-ping"
        style={{ animationDuration: "1.5s" }}
      />
    </div>
  )
}
