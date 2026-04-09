import { type ReactNode } from "react"
import { useQuery } from "@tanstack/react-query"
import { Link, Outlet } from "@tanstack/react-router"
import { AuthProvider, useAuth } from "../../contexts/AuthContext"
import { Button } from "../ui"
import { RootFolderWarning } from "../common"
import { LoginPage } from "../auth"

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
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-2xl font-bold text-foreground">📚 Bookaneer</h1>
          <div className="flex items-center gap-4">
            {health.isLoading ? (
              <span className="text-muted-foreground">Checking...</span>
            ) : health.data?.status === "ok" ? (
              <span className="inline-flex items-center gap-1 text-green-600">
                <span className="h-2 w-2 rounded-full bg-green-500" />
                Connected
              </span>
            ) : (
              <span className="text-destructive">Disconnected</span>
            )}
            {user && (
              <span className="text-sm text-muted-foreground">
                {user.username || "API Key"}
              </span>
            )}
            <Button variant="outline" size="sm" onClick={logout}>
              Sign Out
            </Button>
          </div>
        </div>
        <Navigation />
      </header>

      <main className="container mx-auto px-4 py-8">
        <RootFolderWarning />
        {children}
      </main>
    </div>
  )
}

// Navigation component using TanStack Router Link
function Navigation() {
  const navItems = [
    { to: "/", label: "Library" },
    { to: "/authors", label: "Authors" },
    { to: "/books", label: "Books" },
    { to: "/wanted", label: "Wanted" },
    { to: "/activity", label: "Activity" },
    { to: "/search", label: "Search" },
    { to: "/settings", label: "Settings" },
    { to: "/system", label: "System" },
  ] as const

  return (
    <nav className="container mx-auto px-4" aria-label="Main navigation">
      <ul className="flex gap-1 -mb-px" role="tablist">
        {navItems.map((item) => (
          <li key={item.to} role="presentation">
            <Link
              to={item.to}
              className="inline-flex items-center justify-center px-4 py-2 text-sm font-medium rounded-none border-b-2 transition-colors"
              activeProps={{
                className: "border-primary text-primary",
              }}
              inactiveProps={{
                className: "border-transparent text-muted-foreground hover:text-foreground",
              }}
              activeOptions={{ exact: item.to === "/" }}
            >
              {item.label}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
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
