---
applyTo: "**/*.tsx,**/*.jsx"
---

# React Conventions

## Components
- Functional components with hooks — no class components
- Named exports (no default exports)
- Explicit `Props` interface for every component
- Keep components focused — split large components into smaller ones

## State Management
- **Server state**: TanStack Query (React Query) — queries with descriptive keys
- **UI state**: `useState` for local, Zustand for shared feature state
- **URL state**: Router params and search params
- **Complex flows**: XState for multi-step state machines

## Hooks
- Custom hooks for data fetching: `useBooks()`, `useCreateBook()`
- Wrap TanStack Query calls in custom hooks — never call `useQuery` directly in components
- Invalidate queries after mutations

## Accessibility
- Semantic HTML elements (`<button>`, `<nav>`, `<main>`, not `<div onClick>`)
- `aria-*` attributes where semantic HTML is insufficient
- Keyboard navigation support
- Proper heading hierarchy (`h1` → `h2` → `h3`)

## Patterns
- Colocate related files: `BookCard.tsx` + `BookCard.test.tsx` + `useBookData.ts`
- Feature-sliced structure for large apps: `features/featureName/{api,components,hooks,types}/`
- API calls through a shared client (axios or fetch wrapper)
