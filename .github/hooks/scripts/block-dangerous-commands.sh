#!/usr/bin/env bash
# block-dangerous-commands.sh
#
# Copilot hook script for PreToolUse events.
# Reads the tool invocation from stdin (JSON), checks the command
# against rules in guardrails-rules.txt, and returns a permissionDecision.
#
# Exit codes:
#   0 = allowed (or tool is not a terminal command)
#
# Output (JSON to stdout):
#   {"permissionDecision": "allow"}
#   {"permissionDecision": "deny", "reason": "..."}
#   {"permissionDecision": "ask", "reason": "..."}

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RULES_FILE="${SCRIPT_DIR}/../guardrails-rules.txt"

# Read the hook payload from stdin
PAYLOAD=$(cat)

# Extract the tool name and command
TOOL_NAME=$(echo "$PAYLOAD" | grep -o '"toolName"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"toolName"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
COMMAND=$(echo "$PAYLOAD" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"command"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

# Only check terminal commands
if [[ "$TOOL_NAME" != *"terminal"* && "$TOOL_NAME" != *"shell"* && "$TOOL_NAME" != "runInTerminal" ]]; then
  echo '{"permissionDecision": "allow"}'
  exit 0
fi

# If no command found, allow
if [[ -z "$COMMAND" ]]; then
  echo '{"permissionDecision": "allow"}'
  exit 0
fi

# If rules file doesn't exist, allow everything
if [[ ! -f "$RULES_FILE" ]]; then
  echo '{"permissionDecision": "allow"}'
  exit 0
fi

# Check command against each rule
while IFS= read -r line; do
  # Skip comments and blank lines
  [[ "$line" =~ ^[[:space:]]*# ]] && continue
  [[ "$line" =~ ^[[:space:]]*$ ]] && continue

  # Parse action and pattern
  ACTION="${line%%[[:space:]]*}"
  PATTERN="${line#*[[:space:]]}"
  # Trim leading whitespace from pattern
  PATTERN="${PATTERN#"${PATTERN%%[![:space:]]*}"}"

  # Skip malformed lines
  [[ -z "$ACTION" || -z "$PATTERN" ]] && continue

  # Check if command matches pattern
  if echo "$COMMAND" | grep -qiE "$PATTERN"; then
    case "$ACTION" in
      BLOCK)
        echo "{\"permissionDecision\": \"deny\", \"reason\": \"Blocked by guardrail: pattern '$PATTERN' matched\"}"
        exit 0
        ;;
      ASK)
        echo "{\"permissionDecision\": \"ask\", \"reason\": \"Requires confirmation: pattern '$PATTERN' matched\"}"
        exit 0
        ;;
    esac
  fi
done < "$RULES_FILE"

# No rules matched — allow
echo '{"permissionDecision": "allow"}'
exit 0
