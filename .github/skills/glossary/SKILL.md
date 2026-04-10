---
name: glossary
description: "Extract and formalize domain terminology into a glossary. Use when: defining domain terms, building a glossary, hardening terminology, domain model, DDD, ubiquitous language, glossary."
---

# Glossary Builder

Extract domain terminology from conversation, specs, or code and formalize it
into a canonical glossary saved as `GLOSSARY.md`.

## When to Use

- Starting a new project and aligning on terminology
- Domain terms are used inconsistently across the codebase
- Multiple people use different words for the same concept
- Onboarding new team members who need a term reference

## Procedure

1. Scan the conversation, specs, or code for domain-relevant nouns, verbs, and concepts
2. Identify problems:
   - Same word used for different concepts (ambiguity)
   - Different words used for the same concept (synonyms)
   - Vague or overloaded terms
3. Propose a canonical glossary — be **opinionated** about term choices
4. Write to `GLOSSARY.md` in the project root using the [template](./assets/glossary-template.md)
5. If `GLOSSARY.md` already exists, read it first and incorporate new terms

## Rules

- **Be opinionated** — pick the best term, list others as "aliases to avoid"
- **Flag ambiguities explicitly** — call out conflicts with clear recommendations
- **Domain terms only** — skip generic programming concepts (array, function, endpoint)
- **Tight definitions** — one sentence max; define what it IS, not what it does
- **Show relationships** — use bold term names, express cardinality where obvious
- **Group by context** — split into tables when natural clusters emerge
- **Include example dialogue** — 3-5 exchanges showing terms used precisely
