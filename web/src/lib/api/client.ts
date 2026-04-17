import * as z from "zod"

const API_BASE = "/api/v1"
const API_KEY_STORAGE_KEY = "bookaneer_api_key"

const storedKeySchema = z.string().min(1)

// Get stored API key (validated — rejects empty strings)
export function getStoredApiKey(): string | null {
  const raw = localStorage.getItem(API_KEY_STORAGE_KEY)
  const result = storedKeySchema.safeParse(raw)
  return result.success ? result.data : null
}

// Set stored API key
export function setStoredApiKey(apiKey: string): void {
  localStorage.setItem(API_KEY_STORAGE_KEY, apiKey)
}

// Clear stored API key
export function clearStoredApiKey(): void {
  localStorage.removeItem(API_KEY_STORAGE_KEY)
}

// Generic fetch wrapper with auth and optional Zod schema validation
export async function fetchAPI<T>(
  path: string,
  options?: RequestInit,
  schema?: z.ZodSchema<T>,
): Promise<T> {
  const apiKey = getStoredApiKey()
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options?.headers,
  }
  if (apiKey) {
    ;(headers as Record<string, string>)["X-Api-Key"] = apiKey
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: res.statusText }))
    throw new Error(error.message || res.statusText)
  }

  // Handle 204 No Content or empty responses
  if (res.status === 204 || res.headers.get("content-length") === "0") {
    return undefined as T
  }

  const data = await res.json()
  if (schema) {
    return schema.parse(data)
  }
  return data as T
}

export { API_BASE }
