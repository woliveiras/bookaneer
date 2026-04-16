import { Globe, Library } from "lucide-react"

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
  .animate-sail { animation: sail 4s ease-in-out infinite; }
  .animate-bob { animation: bob 2s ease-in-out infinite; }
  .animate-waves { animation: waves 3s linear infinite; }
`

export function SearchLoadingAnimation() {
  return (
    <div className="py-8">
      <div className="relative h-24 mx-auto max-w-sm overflow-hidden rounded-lg">
        <div className="absolute bottom-4 left-2" title="Library">
          <Library className="w-6 h-6" />
        </div>
        <div className="absolute bottom-4 right-2" title="Your Collection">
          <Globe className="w-6 h-6" />
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
        <p className="text-sm text-muted-foreground animate-pulse">
          Sailing the seven seas for books...
        </p>
      </div>

      <style>{searchAnimationStyles}</style>
    </div>
  )
}
