---
name: spec-template
description: "Generate a structured spec from direct input. Use when: you already know what you want and just need the formatted template. For exploration and interviews, use the spec-writer agent instead."
---

# Spec Template

Generate a structured feature specification using the template below.
Use this when you already know the requirements and just need them formatted.

## When to Use

- You have a clear idea of what to build
- You want a formatted spec without an interview
- You're documenting an existing decision

For unclear requirements that need exploration, use the `spec-writer` agent instead.

## Procedure

1. Ask the user for the feature name and a brief description
2. Fill in the [spec template](./assets/spec-template.md)
3. Save to `specs/<feature-name>.md`

## Output

Use the template from [assets/spec-template.md](./assets/spec-template.md).
Ensure every acceptance criterion is testable (maps to at least one unit test).
