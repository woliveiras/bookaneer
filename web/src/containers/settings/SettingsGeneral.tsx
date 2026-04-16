import { useState } from "react"
import { Button } from "../../components/ui/Button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../../components/ui/Card"
import { Input } from "../../components/ui/Input"
import { useGeneralSettings } from "../../hooks/useSettings"

export function SettingsGeneral() {
  const { data: settings, isLoading, error } = useGeneralSettings()
  const [showApiKey, setShowApiKey] = useState(false)
  const [copied, setCopied] = useState(false)

  async function copyApiKey() {
    if (!settings?.apiKey) return
    try {
      await navigator.clipboard.writeText(settings.apiKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Fallback for older browsers
      const textArea = document.createElement("textarea")
      textArea.value = settings.apiKey
      document.body.appendChild(textArea)
      textArea.select()
      document.execCommand("copy")
      document.body.removeChild(textArea)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 text-destructive bg-destructive/10 rounded-md">
        Failed to load settings: {error.message}
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Security</CardTitle>
          <CardDescription>
            API key for external integrations (OPDS readers, scripts, automation)
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <label htmlFor="api-key" className="block text-sm font-medium mb-1.5">
              API Key
            </label>
            <div className="flex flex-col sm:flex-row gap-2">
              <div className="relative flex-1">
                <Input
                  id="api-key"
                  type={showApiKey ? "text" : "password"}
                  value={settings?.apiKey || ""}
                  readOnly
                  className="font-mono pr-20"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-1 top-1/2 -translate-y-1/2 h-7 text-xs"
                  onClick={() => setShowApiKey(!showApiKey)}
                >
                  {showApiKey ? "Hide" : "Show"}
                </Button>
              </div>
              <Button type="button" variant="outline" onClick={copyApiKey}>
                {copied ? "Copied!" : "Copy"}
              </Button>
            </div>
            <p className="mt-1.5 text-xs text-muted-foreground">
              Use this key for external applications. Include it in the X-Api-Key header or ?apikey=
              query parameter.
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Application</CardTitle>
          <CardDescription>General application configuration (read-only)</CardDescription>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <dt className="text-muted-foreground">Bind Address</dt>
              <dd className="font-mono">
                {settings?.bindAddress}:{settings?.port}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Log Level</dt>
              <dd className="font-mono">{settings?.logLevel}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Data Directory</dt>
              <dd className="font-mono break-all">{settings?.dataDir}</dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Library Directory</dt>
              <dd className="font-mono break-all">{settings?.libraryDir}</dd>
            </div>
          </dl>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Custom Providers</CardTitle>
          <CardDescription>
            Configured custom providers currently active in this instance
          </CardDescription>
        </CardHeader>
        <CardContent>
          {!settings?.customProvidersEnabled ? (
            <p className="text-sm text-muted-foreground">
              Custom providers are disabled (`customProvidersEnabled: false`).
            </p>
          ) : settings.customProvidersActive.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              Custom providers are enabled, but no valid active providers were found.
            </p>
          ) : (
            <div className="space-y-2">
              {settings.customProvidersActive.map((provider) => (
                <div
                  key={`${provider.name}-${provider.domain}`}
                  className="rounded-md border p-3 text-sm"
                >
                  <div>
                    <span className="text-muted-foreground">Name: </span>
                    <span className="font-medium">{provider.name || "(missing)"}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">File type: </span>
                    <span className="font-mono">{provider.formatHint || "(auto)"}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Domain: </span>
                    <span className="text-muted-foreground font-mono break-all">
                      {provider.domain || "(missing)"}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
