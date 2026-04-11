// Ship SVG component for search loading animation
export function Ship({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 64 48" className={className} fill="none" xmlns="http://www.w3.org/2000/svg">
      <title>Ship</title>
      <path
        d="M8 32 L12 40 L52 40 L56 32 L48 32 L48 28 L16 28 L16 32 Z"
        fill="currentColor"
        className="text-amber-800"
      />
      <rect x="16" y="24" width="32" height="4" fill="currentColor" className="text-amber-700" />
      <rect x="30" y="4" width="4" height="24" fill="currentColor" className="text-amber-900" />
      <path d="M34 6 L34 22 L50 22 Q42 14 34 6 Z" fill="currentColor" className="text-slate-100" />
      <rect x="31" y="0" width="12" height="8" fill="currentColor" className="text-slate-800" />
    </svg>
  )
}

export function WavesSVG() {
  return (
    <svg
      viewBox="0 0 400 20"
      className="absolute bottom-0 left-0 w-[200%] h-5"
      preserveAspectRatio="none"
    >
      <title>Ocean waves</title>
      <path
        d="M0 10 Q25 0 50 10 T100 10 T150 10 T200 10 T250 10 T300 10 T350 10 T400 10 V20 H0 Z"
        fill="currentColor"
        className="text-blue-400/40"
      />
      <path
        d="M0 14 Q25 8 50 14 T100 14 T150 14 T200 14 T250 14 T300 14 T350 14 T400 14 V20 H0 Z"
        fill="currentColor"
        className="text-blue-500/50"
      />
    </svg>
  )
}

export const searchAnimationStyles = `
  @keyframes sail { 0%, 100% { left: 10%; } 50% { left: calc(90% - 4rem); } }
  @keyframes bob { 0%, 100% { transform: translateY(0) rotate(-2deg); } 25% { transform: translateY(-3px) rotate(0deg); } 50% { transform: translateY(0) rotate(2deg); } 75% { transform: translateY(2px) rotate(0deg); } }
  @keyframes waves { 0% { transform: translateX(0); } 100% { transform: translateX(-50%); } }
  @keyframes gradient-spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
  .animate-sail { animation: sail 4s ease-in-out infinite; }
  .animate-bob { animation: bob 2s ease-in-out infinite; }
  .animate-waves { animation: waves 3s linear infinite; }
  .animate-gradient-spin { animation: gradient-spin 1s linear infinite; }
`

interface SearchSourceStatus {
  name: string
  done: boolean
  error: unknown
  retrying: boolean
}

interface SearchLoadingAnimationProps {
  sources: SearchSourceStatus[]
}

export function SearchLoadingAnimation({ sources }: SearchLoadingAnimationProps) {
  return (
    <div className="py-6">
      <div className="relative h-24 mx-auto max-w-sm overflow-hidden rounded-lg">
        <div className="absolute bottom-4 left-2 text-2xl" title="Library">
          📚
        </div>
        <div className="absolute bottom-4 right-2 text-2xl" title="Your Collection">
          🌍
        </div>
        <div className="absolute bottom-3 w-16 h-12 animate-sail">
          <div className="animate-bob">
            <Ship className="w-full h-full drop-shadow-md" />
          </div>
        </div>
        <div className="absolute bottom-0 left-0 w-full overflow-hidden animate-waves">
          <WavesSVG />
        </div>
      </div>
      <div className="text-center mt-4">
        {sources.some((s) => s.retrying) ? (
          <p className="text-sm text-amber-500 animate-pulse">
            Some sources had issues, retrying...
          </p>
        ) : (
          <p className="text-sm text-muted-foreground animate-pulse">
            Sailing the seven seas for books...
          </p>
        )}
      </div>
      <div className="flex justify-center gap-3 mt-3 flex-wrap">
        {sources.map((source) => (
          <div
            key={source.name}
            className="flex items-center gap-1.5 text-xs text-muted-foreground"
          >
            {source.done ? (
              source.error ? (
                <span className="text-destructive">✗</span>
              ) : (
                <span className="text-green-500">✓</span>
              )
            ) : (
              <div className="relative h-3.5 w-3.5">
                <div
                  className="absolute inset-0 rounded-full animate-gradient-spin"
                  style={{
                    background: source.retrying
                      ? "conic-gradient(from 0deg, transparent, #f59e0b, #eab308, transparent)"
                      : "conic-gradient(from 0deg, transparent, #60a5fa, #3b82f6, transparent)",
                  }}
                />
                <div className="absolute inset-0.5 rounded-full bg-background" />
              </div>
            )}
            <span className="hidden sm:inline">
              {source.name}
              {source.retrying && <span className="text-amber-500 ml-1">(retrying...)</span>}
            </span>
          </div>
        ))}
      </div>

      <style>{searchAnimationStyles}</style>
    </div>
  )
}
