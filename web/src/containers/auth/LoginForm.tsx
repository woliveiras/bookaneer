import { Flag } from "lucide-react"
import { type FormEvent, useState } from "react"
import { useSelector } from "@xstate/react"
import * as z from "zod"
import { Button } from "../../components/ui/Button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card"
import { Input } from "../../components/ui/Input"
import { useAuthActor } from "../../features/auth/AuthProvider"

const loginSchema = z.object({
  username: z
    .string()
    .min(1, { error: "Username is required" })
    .transform((s) => s.trim()),
  password: z.string().min(1, { error: "Password is required" }),
})

export function LoginForm() {
  const actorRef = useAuthActor()
  const error = useSelector(actorRef, (s) => s.context.error)
  const isLoading = useSelector(actorRef, (s) => s.matches("loggingIn"))
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [validationError, setValidationError] = useState<string | null>(null)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setValidationError(null)

    const result = loginSchema.safeParse({ username, password })
    if (!result.success) {
      setValidationError(z.prettifyError(result.error))
      return
    }

    actorRef.send({
      type: "LOGIN_WITH_CREDENTIALS",
      username: result.data.username,
      password: result.data.password,
    })
  }

  const displayError = validationError ?? error

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="text-center">
        <div className="flex items-center justify-center gap-2 mb-2">
          <Flag className="w-8 h-8" />
          <CardTitle className="text-2xl">Bookaneer</CardTitle>
        </div>
        <CardDescription>Sign in to access your library</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="username-input" className="block text-sm font-medium mb-1.5">
              Username
            </label>
            <Input
              id="username-input"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="admin"
              autoComplete="username"
              autoFocus
            />
          </div>
          <div>
            <label htmlFor="password-input" className="block text-sm font-medium mb-1.5">
              Password
            </label>
            <Input
              id="password-input"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              autoComplete="current-password"
            />
          </div>

          {displayError && (
            <div
              className="p-3 text-sm text-destructive bg-destructive/10 rounded-md"
              role="alert"
              aria-live="polite"
            >
              {displayError}
            </div>
          )}

          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "Signing in..." : "Sign In"}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
