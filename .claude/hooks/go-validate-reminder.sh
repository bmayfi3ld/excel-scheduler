#!/bin/sh
# Claude Code PostToolUse hook.
#
# After an Edit/Write/MultiEdit touches a Go (.go) file, remind Claude to run
# `just validate` (unit tests + golangci-lint) before committing. Edits to any
# other file type produce no output and the hook is a no-op.
#
# Wired up in .claude/settings.json under hooks.PostToolUse.

input=$(cat)

if command -v jq >/dev/null 2>&1; then
  file=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')
else
  # jq-less fallback: pull the first "file_path": "..." value out of the JSON.
  file=$(printf '%s' "$input" | sed -n 's/.*"file_path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)
fi

case "$file" in
  *.go)
    printf '%s\n' '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"You edited a Go file. Run `just validate` (Go unit tests + golangci-lint security/complexity checks) before committing to confirm the change passes."}}'
    ;;
esac

exit 0
