---
description: "Use when writing, reviewing, or refactoring Tailwind CSS utility classes. Covers Tailwind v4 syntax, CSS-first configuration, theme variables, breaking changes from v3, and patterns to avoid common linter/build errors."
applyTo: "web/**/*.tsx, web/**/*.ts, web/**/*.css"
---

# Tailwind CSS v4 — Conventions and Patterns

This project uses **Tailwind CSS v4** (`tailwindcss ^4.2.x`) with the **Vite plugin** (`@tailwindcss/vite`).
v4 is a ground-up rewrite. Many patterns from v3 are invalid or deprecated.

---

## Setup (project-specific)

```css
/* web/src/index.css */
@import "tailwindcss";   /* single import — no @tailwind directives */

@theme {
  /* design tokens go here */
}
```

```ts
// vite.config.ts
import tailwindcss from "@tailwindcss/vite"
export default defineConfig({ plugins: [tailwindcss()] })
```

- **No `tailwind.config.js`** — configuration lives in CSS via `@theme`
- **No `content` array** — Tailwind auto-detects source files (excludes `.gitignore`, `node_modules`, binary files, CSS files)
- **No `postcss-import` or `autoprefixer`** — handled automatically

---

## CSS-First Configuration (`@theme`)

All design tokens are defined in CSS, not JS:

```css
@import "tailwindcss";

@theme {
  --color-brand: oklch(0.72 0.11 221);
  --font-display: "Inter", sans-serif;
  --breakpoint-3xl: 120rem;
  --radius-card: 0.75rem;
}
```

- Each `--color-*` var creates utilities like `bg-brand`, `text-brand`, `border-brand`
- Each `--font-*` creates `font-display`
- Each `--breakpoint-*` creates a responsive variant `3xl:*`
- **Use `@theme` for design tokens** that should generate utilities
- **Use `:root` for regular CSS variables** that should NOT generate utilities

### Extending vs Overriding

```css
/* Extend — keep defaults, add new */
@theme {
  --color-mint: oklch(0.72 0.11 178);
}

/* Override entire namespace — clear all defaults first */
@theme {
  --color-*: initial;
  --color-white: #fff;
  --color-primary: oklch(0.14 0.005 286);
}
```

---

## ⚠️ Breaking Changes from v3 — Most Common Mistakes

### 1. Import syntax

```css
/* ❌ v3 — does not work in v4 */
@tailwind base;
@tailwind components;
@tailwind utilities;

/* ✅ v4 */
@import "tailwindcss";
```

### 2. Renamed utilities (shadow, blur, rounded)

The scale shifted — the "bare" names changed meaning:

| v3 class | v4 class |
|----------|----------|
| `shadow-sm` | `shadow-xs` |
| `shadow` | `shadow-sm` |
| `blur-sm` | `blur-xs` |
| `blur` | `blur-sm` |
| `rounded-sm` | `rounded-xs` |
| `rounded` | `rounded-sm` |
| `drop-shadow-sm` | `drop-shadow-xs` |
| `drop-shadow` | `drop-shadow-sm` |

### 3. Removed deprecated utilities

```css
/* ❌ Removed in v4 */
bg-opacity-50     → bg-black/50
text-opacity-50   → text-black/50
border-opacity-50 → border-black/50
ring-opacity-50   → ring-black/50
flex-shrink-0     → shrink-0
flex-grow         → grow

/* ✅ Use opacity modifiers */
<div class="bg-black/50 text-white/75 border-red-500/30">
```

### 4. `outline-none` → `outline-hidden`

```html
<!-- ❌ v3 behavior: invisible outline (accessibility-safe) -->
<input class="focus:outline-none" />

<!-- ✅ v4: same accessible behavior -->
<input class="focus:outline-hidden" />

<!-- v4 outline-none now literally sets outline-style: none — skips forced-colors mode -->
```

### 5. Default `ring` width changed: 3px → 1px

```html
<!-- ❌ v3: ring = 3px blue -->
<button class="focus:ring" />

<!-- ✅ v4: be explicit -->
<button class="focus:ring-3 focus:ring-blue-500" />
```

### 6. Default border color changed

In v3, `border` used `gray-200` by default. In v4 it uses `currentColor`.
Always specify a border color explicitly:

```html
<!-- ❌ Will not look like gray in v4 -->
<div class="border">

<!-- ✅ Always specify color -->
<div class="border border-border">
<div class="border border-gray-200">
```

### 7. CSS variable arbitrary values syntax

```html
<!-- ❌ v3 syntax (still works but deprecated) -->
<div class="bg-[--brand-color]">

<!-- ✅ v4 syntax -->
<div class="bg-(--brand-color)">

<!-- full var() still works -->
<div class="bg-[var(--brand-color)]">
```

### 8. `!important` modifier moved to the end

```html
<!-- ❌ v3 (deprecated) -->
<div class="!flex !bg-red-500">

<!-- ✅ v4 -->
<div class="flex! bg-red-500!">
```

### 9. Variant stacking order reversed (left-to-right)

```html
<!-- ❌ v3 (right-to-left) -->
<ul class="first:*:pt-0 last:*:pb-0">

<!-- ✅ v4 (left-to-right — reads like CSS selector) -->
<ul class="*:first:pt-0 *:last:pb-0">
```

### 10. Gradients — partial overrides no longer reset

```html
<!-- v3: dark:from-blue-500 would reset via/to to transparent -->
<!-- v4: all stops are preserved — use via-none to explicitly reset -->
<div class="bg-linear-to-r from-red-500 via-orange-400 to-yellow-400 dark:via-none dark:from-blue-500 dark:to-teal-400">
```

### 11. `transform-none` no longer resets individual transforms

```html
<!-- ❌ v4: transform-none won't reset scale-* -->
<button class="scale-150 focus:transform-none">

<!-- ✅ v4 -->
<button class="scale-150 focus:scale-none">
```

### 12. `hover:` now requires pointer device

v4 wraps `hover:` in `@media (hover: hover)` automatically.
Touch devices won't trigger hover on tap. If you need the old behavior:

```css
/* index.css — override hover to always apply */
@custom-variant hover (&:hover);
```

---

## Custom Utilities and Components

```css
/* ❌ v3 — @layer utilities is now a native CSS layer, not Tailwind magic */
@layer utilities {
  .tab-4 { tab-size: 4; }
}

/* ✅ v4 — use @utility for Tailwind-integrated utilities */
@utility tab-4 {
  tab-size: 4;
}

/* ✅ base styles still use @layer base */
@layer base {
  h1 { font-size: var(--text-2xl); }
}
```

---

## Dynamic Class Names — Important Rule

Tailwind scans files as plain text. **Class names must be complete strings** — never construct them dynamically:

```tsx
// ❌ Tailwind will NOT detect these
<div className={`text-${color}-600`} />
<div className={`bg-${size === 'sm' ? 'small' : 'large'}`} />

// ✅ Always use complete class names
const colorMap = {
  red: "text-red-600",
  green: "text-green-600",
}
<div className={colorMap[color]} />
```

---

## Container Queries (built-in, no plugin)

```html
<!-- Parent establishes container context -->
<div class="@container">
  <div class="@sm:grid-cols-3 grid grid-cols-1">
    <!-- Responds to container width, not viewport -->
  </div>
</div>
```

---

## New Utilities in v4

| Utility | What it does |
|---------|-------------|
| `bg-linear-to-r` | Replaces `bg-gradient-to-r` (still works) |
| `@container` | Container query context |
| `@sm:*`, `@md:*` | Container query variants |
| `starting:*` | `@starting-style` — enter/exit transitions without JS |
| `not-*` | Negate a variant: `not-hover:opacity-50` |
| `inert` | Style inert elements |
| `field-sizing-content` | Auto-resize textarea |
| `rotate-x-*`, `rotate-y-*` | 3D transforms |
| `perspective-*` | 3D perspective |
| `bg-radial-*`, `bg-conic-*` | Radial and conic gradients |

---

## Source Detection — When to Use `@source`

```css
/* Scan a library ignored by .gitignore */
@import "tailwindcss";
@source "../node_modules/@my-company/ui-lib";

/* Safelist classes not found in source */
@source inline("underline");
@source inline("{hover:,focus:,}ring-2");

/* Exclude a large legacy directory */
@source not "../src/legacy";
```

---

## Prefix Usage (if ever needed)

```css
@import "tailwindcss" prefix(tw);
```

```html
<div class="tw:flex tw:bg-red-500 tw:hover:bg-red-600">
```

---

## Biome / Linter Checklist

Common lint errors specific to this codebase:

- `noUnusedImports` — remove unused React/lucide imports after refactoring
- `noNoninteractiveElementToInteractiveRole` — don't add `role="tablist"` to `<ul>`, use a `<div>` or remove the role
- `useValidAriaValues` — `aria-current` only accepts: `false | true | "page" | "step" | "location" | "date" | "time"` — do not pass `undefined`
- Always use `aria-hidden="true"` on decorative icons (`<Icon aria-hidden="true" />`)
