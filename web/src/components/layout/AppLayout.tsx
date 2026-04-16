import { useQuery } from "@tanstack/react-query"
import { Link, Outlet } from "@tanstack/react-router"
import {
  Activity,
  BookOpen,
  Library,
  Menu,
  Search,
  Settings,
  Star,
  Users,
  X,
} from "lucide-react"
import { type ReactNode, useState } from "react"
import { AuthProvider, useAuth } from "../../contexts/AuthContext"
import { LoginPage } from "../../pages/LoginPage"
import { RootFolderWarning } from "../common"
import { Button } from "../ui"

interface HealthResponse {
  status: string
}

// Auth-protected layout wrapper
export function AuthLayout({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading: authLoading, logout, user } = useAuth()

  const health = useQuery<HealthResponse>({
    queryKey: ["health"],
    queryFn: () => fetch("/api/v1/system/health").then((r) => r.json()),
    enabled: isAuthenticated,
  })

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <LoginPage />
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <header className="border-b border-border sticky top-0 z-40 bg-background">
        <div className="container mx-auto px-4 py-3 sm:py-4 flex items-center justify-between gap-2">
          <h1 className="text-xl sm:text-2xl font-bold text-foreground whitespace-nowrap flex items-center gap-2">
            <Library className="w-5 h-5" aria-hidden="true" /> Bookaneer
          </h1>
          <div className="flex items-center gap-2 sm:gap-4">
            {health.isLoading ? (
              <span className="text-muted-foreground text-xs sm:text-sm">Checking...</span>
            ) : health.data?.status === "ok" ? (
              <span className="inline-flex items-center gap-1 text-green-600 text-xs sm:text-sm">
                <span className="h-2 w-2 rounded-full bg-green-500" aria-hidden="true" />
                <span className="hidden sm:inline">Connected</span>
              </span>
            ) : (
              <span className="text-destructive text-xs sm:text-sm">Disconnected</span>
            )}
            {user && (
              <span className="hidden sm:inline text-sm text-muted-foreground">
                {user.username || "API Key"}
              </span>
            )}
            <Button variant="outline" size="sm" onClick={logout}>
              Sign Out
            </Button>
          </div>
        </div>
        {/* Desktop tab navigation — hidden on mobile */}
        <DesktopNavigation />
      </header>

      {/* Main content — extra bottom padding on mobile to clear the bottom nav */}
      <main className="container mx-auto px-4 py-4 sm:py-8 pb-20 sm:pb-8 flex-1">
        <RootFolderWarning />
        {children}
      </main>

      {/* Mobile bottom navigation — visible only below sm breakpoint */}
      <MobileBottomNav />
    </div>
  )
}

// ─── Desktop horizontal tab bar (sm+) ───────────────────────────────────────

const ALL_NAV_ITEMS = [
  { to: "/", label: "Library" },
  { to: "/authors", label: "Authors" },
  { to: "/books", label: "Books" },
  { to: "/wanted", label: "Wishlist" },
  { to: "/activity", label: "Activity" },
  { to: "/search", label: "Search" },
  { to: "/settings", label: "Settings" },
  { to: "/system", label: "System" },
] as const

function DesktopNavigation() {
  return (
    <nav
      className="hidden sm:block container mx-auto px-4"
      aria-label="Main navigation"
    >
      <div className="-mb-px flex flex-wrap gap-1" role="tablist">
        {ALL_NAV_ITEMS.map((item) => (
          <div key={item.to} role="presentation">
            <Link
              to={item.to}
              className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium rounded-none border-b-2 transition-colors whitespace-nowrap border-none"
              activeProps={{ className: "border-primary text-primary" }}
              inactiveProps={{
                className: "border-transparent text-muted-foreground hover:text-foreground",
              }}
              activeOptions={{ exact: item.to === "/" }}
            >
              {item.label}
            </Link>
          </div>
        ))}
      </div>
    </nav>
  )
}

// ─── Mobile bottom navigation (< sm) ────────────────────────────────────────

const PRIMARY_NAV = [
  { to: "/", label: "Library", icon: Library },
  { to: "/books", label: "Books", icon: BookOpen },
  { to: "/authors", label: "Authors", icon: Users },
  { to: "/wanted", label: "Wishlist", icon: Star },
  { to: "/activity", label: "Activity", icon: Activity },
] as const

const MORE_NAV = [
  { to: "/search", label: "Search", icon: Search },
  { to: "/settings", label: "Settings", icon: Settings },
  { to: "/system", label: "System", icon: Library },
] as const

function MobileBottomNav() {
  const [drawerOpen, setDrawerOpen] = useState(false)

  return (
    <>
      {/* Overlay */}
      {drawerOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/40 sm:hidden"
          onClick={() => setDrawerOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* More drawer */}
      {drawerOpen && (
        <div
          className="fixed bottom-16 left-0 right-0 z-50 sm:hidden bg-background border-t border-border shadow-lg rounded-t-xl px-4 py-4"
          role="dialog"
          aria-label="More navigation options"
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-semibold text-foreground">More</span>
            <button
              type="button"
              onClick={() => setDrawerOpen(false)}
              className="p-2 -mr-2 text-muted-foreground hover:text-foreground"
              aria-label="Close menu"
            >
              <X className="w-5 h-5" aria-hidden="true" />
            </button>
          </div>
          <nav aria-label="More navigation">
            <ul className="space-y-1 list-none p-0 m-0">
              {MORE_NAV.map(({ to, label, icon: Icon }) => (
                <li key={to}>
                  <Link
                    to={to}
                    onClick={() => setDrawerOpen(false)}
                    className="flex items-center gap-3 px-3 py-3 rounded-lg text-sm font-medium transition-colors text-muted-foreground hover:text-foreground hover:bg-accent"
                    activeProps={{ className: "text-primary bg-accent" }}
                    activeOptions={{ exact: false }}
                  >
                    <Icon className="w-5 h-5" aria-hidden="true" />
                    {label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>
        </div>
      )}

      {/* Bottom nav bar */}
      <nav
        className="fixed bottom-0 left-0 right-0 z-40 sm:hidden bg-background border-t border-border"
        aria-label="Primary navigation"
        style={{ paddingBottom: "env(safe-area-inset-bottom)" }}
      >
        <ul className="flex list-none p-0 m-0">
          {PRIMARY_NAV.map(({ to, label, icon: Icon }) => (
            <li key={to} className="flex-1" role="presentation">
              <Link
                to={to}
                className="flex flex-col items-center justify-center gap-0.5 py-2 min-h-[56px] w-full text-xs font-medium transition-colors text-muted-foreground"
                activeProps={{ className: "text-primary" }}
                inactiveProps={{ className: "text-muted-foreground" }}
                activeOptions={{ exact: to === "/" }}
              >
                <Icon className="w-5 h-5" aria-hidden="true" />
                <span>{label}</span>
              </Link>
            </li>
          ))}
          {/* More button */}
          <li className="flex-1" role="presentation">
            <button
              type="button"
              onClick={() => setDrawerOpen((v) => !v)}
              className="flex flex-col items-center justify-center gap-0.5 py-2 min-h-[56px] w-full text-xs font-medium transition-colors text-muted-foreground"
              aria-expanded={drawerOpen}
              aria-haspopup="dialog"
            >
              <Menu className="w-5 h-5" aria-hidden="true" />
              <span>More</span>
            </button>
          </li>
        </ul>
      </nav>
    </>
  )
}

// Root layout with auth provider
export function RootLayout() {
  return (
    <AuthProvider>
      <Outlet />
    </AuthProvider>
  )
}
