#!/usr/bin/env bash
# PostToolUse hook: auto-format Go files after edits.
# Receives JSON on stdin with tool invocation details.
# Exits 0 silently if not a Go file or if goimports is not installed.

set -euo pipefail

INPUT=$(cat)

# Extract the file path from the tool input JSON
FILE=$(echo "$INPUT" | grep -o '"filePath"\s*:\s*"[^"]*"' | head -1 | sed 's/.*"filePath"\s*:\s*"//;s/"$//')

# Only run on .go files
if [[ -z "$FILE" || "$FILE" != *.go || ! -f "$FILE" ]]; then
  exit 0
fi

# Prefer goimports, fall back to gofmt
if command -v goimports &>/dev/null; then
  goimports -w "$FILE"
elif command -v gofmt &>/dev/null; then
  gofmt -w "$FILE"
fi
