---
applyTo: "**"
description: "Conventional Commits format for all commit messages."
---

# Conventional Commits

Format: `type(scope): description`

## Types
- `feat` — new feature or capability
- `fix` — bug fix
- `docs` — documentation changes only
- `test` — adding or updating tests
- `refactor` — code change that neither fixes a bug nor adds a feature
- `chore` — maintenance tasks (deps, CI, configs)
- `style` — formatting, whitespace (no code logic change)
- `perf` — performance improvement

## Rules
- Use lowercase for type and description
- Scope is optional but recommended (e.g., `feat(auth): add login`)
- Description is imperative mood: "add", not "added" or "adds"
- Keep the first line under 72 characters
- Breaking changes: add `!` after type — `feat!: remove legacy API`
