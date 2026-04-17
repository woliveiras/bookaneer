---
description: "Use when creating, editing, or reviewing Zod v4 schemas. Covers schema definition, type inference, validation patterns, common v4 API differences, and runtime validation in fetchAPI."
applyTo: "web/src/lib/schemas/**/*.ts, web/src/lib/api/**/*.ts"
---

# Zod v4 — Schema Conventions

## Schema location

All schemas live in `web/src/lib/schemas/`, one file per domain:

```
web/src/lib/schemas/
├── auth.schema.ts
├── book.schema.ts
├── author.schema.ts
├── indexer.schema.ts
├── download.schema.ts
├── naming.schema.ts
├── search.schema.ts
├── settings.schema.ts
├── reader.schema.ts
├── wanted.schema.ts
└── index.ts         # barrel re-exporting all schemas and types
```

Types are **always** inferred from schemas — never define them separately:

```ts
// ✅ Inferred type
export const UserSchema = z.object({ id: z.number(), username: z.string() })
export type User = z.infer<typeof UserSchema>

// ❌ Duplicate manual interface
export interface User { id: number; username: string }
```

## Import style

```ts
import * as z from "zod"
```

## Zod v4 API — key differences from v3

### `z.record` requires two arguments
```ts
// ✅ v4
z.record(z.string(), z.unknown())
z.record(z.string(), z.number())

// ❌ v3-only (fails in v4)
z.record(z.unknown())
```

### Error formatting
```ts
// ✅ v4 — pretty-print for user-facing messages
import * as z from "zod"
const result = schema.safeParse(data)
if (!result.success) {
  const message = z.prettifyError(result.error)
  setError(message)
}

// ❌ v3 — method on error object (removed in v4)
result.error.format()
```

### `.min` / `.max` use `error` key instead of `message`
```ts
// ✅ v4
z.string().min(1, { error: "Required" })
z.number().max(100, { error: "Must be 100 or less" })

// ❌ v3 (still compiles but deprecated)
z.string().min(1, { message: "Required" })
```

## Common patterns

### Optional API fields
```ts
z.string().optional()     // string | undefined
z.string().nullable()     // string | null
z.string().nullish()      // string | null | undefined
```

### Generic response wrapper (e.g. paginated lists)
```ts
export function PaginatedResponseSchema<T extends z.ZodTypeAny>(itemSchema: T) {
  return z.object({
    records: z.array(itemSchema),
    totalRecords: z.number(),
  })
}
// Usage:
PaginatedResponseSchema(BookSchema)
```

### Discriminated unions
```ts
export const ColumnColorHintSchema = z.discriminatedUnion("type", [
  z.object({ type: z.literal("map"),    value: z.string() }),
  z.object({ type: z.literal("static"), value: z.string() }),
])
```

## Runtime validation in `fetchAPI`

Pass a schema as the third argument to `fetchAPI` for runtime validation:

```ts
// Validated — throws ZodError if response shape is wrong
return fetchAPI<Book>("/book/1", undefined, BookSchema)

// Not validated — use for endpoints not yet migrated (backwards-compatible)
return fetchAPI<Book>("/book/1")
```

The `fetchAPI` signature:
```ts
async function fetchAPI<T>(
  path: string,
  options?: RequestInit,
  schema?: z.ZodSchema<T>,
): Promise<T>
```

## Form validation

Use `safeParse` + `z.prettifyError` in form handlers:

```ts
const loginSchema = z.object({
  username: z.string().min(1, { error: "Username is required" }).transform(s => s.trim()),
  password: z.string().min(1, { error: "Password is required" }),
})

const result = loginSchema.safeParse({ username, password })
if (!result.success) {
  setError(z.prettifyError(result.error))
  return
}
// result.data is now typed and safe to use
await login(result.data.username, result.data.password)
```

## `getStoredApiKey` validation

Use `safeParse` to validate values read from `localStorage`:

```ts
const storedKeySchema = z.string().min(1)

export function getStoredApiKey(): string | null {
  const raw = localStorage.getItem(API_KEY_STORAGE_KEY)
  const result = storedKeySchema.safeParse(raw)
  return result.success ? result.data : null
}
```

## Barrel exports

`web/src/lib/schemas/index.ts` re-exports all types and schemas.
`web/src/lib/api.ts` re-exports from `./schemas/index` — all app code imports types
from `../../lib/api` or `../../lib/schemas/<file>` directly.

`lib/types/` was deleted — do NOT recreate it.
