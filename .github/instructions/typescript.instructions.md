---
applyTo: "**/*.ts,**/*.tsx"
---

# TypeScript Conventions

- Enable `strict` mode — no implicit `any`
- Use `interface` for object shapes, `type` for unions and intersections
- Prefer `unknown` over `any`; use type guards to narrow
- Export named exports, not default exports
- Use `const` by default; `let` only when reassignment is needed
- Prefer `readonly` for properties that shouldn't change after creation
- Use template literals over string concatenation
- Handle `null` and `undefined` explicitly — no silent fallbacks
