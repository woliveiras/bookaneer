---
applyTo: "**/*.mdx,**/*.astro"
---

# Astro / MDX Conventions

## Frontmatter
- Every content file requires: `title`, `description`, `pubDate`, `published`, `tags`
- `published: false` hides the post from listings
- Do not repeat the `title` as an H1 in the body — the layout renders it

## Markdown
- Use `##` for main sections, `###` for subsections (no `#` in body)
- Always set a language on code fences: `` ```typescript ``, `` ```bash ``
- Code blocks must be copy-pastable — no placeholder comments
- Use relative paths for internal links

## Content Structure
- One sentence per line (for better diffs)
- 3-7 tags per post
- Keep descriptions to 1-2 sentences (used for OG and listings)

## Custom Embeds
- YouTube: `:youtube[VIDEO_ID]`
- Link cards: `:link[URL]`
- Follow the embed pattern: `embed.ts` + `matcher.ts` + `Component.astro`
