---
description: "Use when writing, reviewing, or refactoring React frontend code. Covers React performance, web accessibility (WCAG), frontend security, TypeScript patterns, component structure, and state management with TanStack Query."
applyTo: "web/**/*.ts, web/**/*.tsx, web/**/*.css"
---

# React Frontend — Performance, Security & Accessibility

## TypeScript

- Strict mode enabled (`"strict": true` in tsconfig)
- No `any` — use `unknown` and narrow with type guards
- Use discriminated unions for state:
  ```ts
  type Result<T> = { status: "success"; data: T } | { status: "error"; error: string }
  ```
- Export types alongside components when they're part of the public API
- Use `satisfies` for type validation without widening: `const config = {...} satisfies Config`

## Component Patterns

- Functional components only — no class components
- Named exports (no default exports): `export function BookCard() {}`
- Co-locate related files: `BookCard.tsx`, `BookCard.test.tsx`, `useBookData.ts`
- Props: explicit interfaces, not inline types
  ```tsx
  interface BookCardProps {
    book: Book
    onSelect: (id: number) => void
  }
  export function BookCard({ book, onSelect }: BookCardProps) {}
  ```
- Prefer composition over prop drilling — use context for deeply nested data

## Performance

### Rendering
- Split large pages into lazy-loaded route chunks via TanStack Router
- Use `React.memo()` only when profiling shows unnecessary re-renders — not preemptively
- Expensive computations: `useMemo` with measured justification
- Stable callbacks for child components: `useCallback` when passing to memoized children
- Never create objects/arrays inline in JSX props: `style={{ color: "red" }}` creates new ref every render

### Data Fetching
- TanStack Query for all server state — no `useEffect` + `useState` for fetching
- Configure `staleTime` per query: library data (30s), system status (5s), config (5min)
- Use `placeholderData` for instant UX while revalidating
- Prefetch on hover for navigation links:
  ```ts
  queryClient.prefetchQuery({ queryKey: ["book", id], queryFn: () => fetchBook(id) })
  ```
- Paginated lists: use `keepPreviousData` to avoid layout shifts

### Images
- Lazy load images below the fold: `loading="lazy"`
- Provide `width` and `height` on `<img>` to prevent layout shift (CLS)
- Use responsive images with `srcSet` when serving multiple cover sizes

### Bundle
- Analyze with `npx vite-bundle-analyzer`
- Dynamic imports for heavy features: `const Reader = lazy(() => import("./features/reader"))`
- Tree-shake: use named imports from libraries, never import entire packages

## Accessibility (WCAG 2.1 AA)

### Semantic HTML
- Use correct elements: `<button>` for actions, `<a>` for navigation, `<nav>`, `<main>`, `<article>`
- Never use `<div>` or `<span>` as interactive elements
- Headings in order: `<h1>` → `<h2>` → `<h3>`, no skipping levels
- Use `<ul>`/`<ol>` for lists (book lists, search results)

### ARIA
- Prefer native semantics over ARIA: a `<button>` doesn't need `role="button"`
- Required ARIA for custom widgets:
  - Modals: `role="dialog"`, `aria-modal="true"`, `aria-labelledby`
  - Tabs: `role="tablist"`, `role="tab"`, `role="tabpanel"`, `aria-selected`
  - Alerts: `role="alert"` for toast notifications
- Status updates: `aria-live="polite"` for async results (search results loaded, download progress)
- Loading states: `aria-busy="true"` on containers being updated

### Keyboard Navigation
- All interactive elements must be focusable and operable via keyboard
- Visible focus indicator on all interactive elements (never `outline: none` without replacement)
- Modal focus trap: Tab cycles within modal, Escape closes
- Skip-to-content link as first focusable element
- Shortcuts: Enter/Space to activate buttons, Escape to close overlays

### Forms
- Every input has a visible `<label>` linked via `htmlFor`/`id`
- Error messages linked to inputs via `aria-describedby`
- Required fields: `aria-required="true"` and visual indicator
- Form validation: show inline errors on blur and on submit

### Color & Contrast
- Minimum contrast ratio 4.5:1 for normal text, 3:1 for large text
- Never convey information by color alone — use icons or text alongside
- Support dark mode and light mode via Tailwind's `dark:` modifier
- Test with browser accessibility tools (Lighthouse, axe DevTools)

### Web Reader Specific
- EPUB reader must support screen reader navigation
- Font size controls: minimum, user-configurable
- High contrast mode for reader
- Keyboard-only navigation through chapters
- Reading progress announced to screen readers

## Security

- Never use `dangerouslySetInnerHTML` — if absolutely necessary, sanitize with DOMPurify
- Sanitize all external metadata (book descriptions, author bios) before rendering
- No `eval()`, `Function()`, or dynamic script loading
- CSP-compatible: no inline styles in JS, use Tailwind classes
- API calls: always include credentials via TanStack Query's `fetch` wrapper
- No secrets in frontend code — all auth via httpOnly cookies or API key header from secure storage
- Validate redirects: only allow relative URLs or allowlisted domains

## State Management

Choose the right tool for each type of state:

| State type | Tool |
|---|---|
| Server state | TanStack Query (books, authors, config, queue) |
| Multi-step flow with loading/error/success | XState v5 (`features/<domain>/<domain>.machine.ts`) |
| Shared UI state across components | Zustand v5 (`store/<domain>/<domain>.store.ts`) |
| Local component state | `useState` / `useReducer` |
| URL state | TanStack Router search params |
| Form state | Controlled components + Zod `safeParse` for validation |

### XState for flows
Use XState when a feature has multiple lifecycle states and side effects:
auth (`checking → authenticated`), reader lifecycle (`idle → loading → ready / failed`),
multi-step wizards, or anything that can be in multiple exclusive states.
See `xstate.instructions.md` for patterns.

### Zustand for shared UI state
Use Zustand when multiple components at different tree levels need to read the same
client-side state: auth credentials, reader preferences, UI configuration.
See `zustand.instructions.md` for patterns.

### XState + Zustand together
XState owns the logic (transitions, guards, side effects).
Zustand is the read cache for components that only need to read the current value.
Sync them in the Provider via `actorRef.subscribe()` — see `AuthProvider.tsx` as reference.

## Styling

- Tailwind CSS utility classes — no custom CSS except for truly unique cases
- shadcn/ui for all base components (Button, Input, Dialog, Table, etc.)
- Responsive: mobile-first (`sm:`, `md:`, `lg:` breakpoints)
- Dark mode: `dark:` variant on all colored elements
- Consistent spacing: use Tailwind's spacing scale, never arbitrary pixel values
