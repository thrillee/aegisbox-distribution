#!/usr/bin/env bash
# PreToolUse hook (Bash tool) -- blocks any git commit/push, or gh pr/issue
# create/edit, that carries AI/Claude attribution text.
#
# Why: explicit, repeated user instruction to never attribute a commit or a
# PR to Claude/AI (see memory feedback_no_ai_attribution.md). This is a
# mechanical backstop on top of that memory -- a fresh session that hasn't
# loaded it yet, or a moment where the default git-commit template habit
# wins out, still gets caught here before the commit/push/PR actually
# happens.
#
# Fails open on anything it can't parse or determine (missing jq/python3,
# malformed stdin, no upstream configured for a push's range check) --
# a broken hook that blocks every Bash call unconditionally would be far
# worse than occasionally missing a violation.
set -uo pipefail

INPUT="$(cat)"

COMMAND=""
CWD=""
if command -v jq >/dev/null 2>&1; then
  COMMAND="$(printf '%s' "$INPUT" | jq -r '.tool_input.command // empty' 2>/dev/null)"
  CWD="$(printf '%s' "$INPUT" | jq -r '.cwd // empty' 2>/dev/null)"
elif command -v python3 >/dev/null 2>&1; then
  COMMAND="$(printf '%s' "$INPUT" | python3 -c 'import json,sys
try:
    d=json.load(sys.stdin)
    print(d.get("tool_input",{}).get("command",""))
except Exception:
    pass' 2>/dev/null)"
  CWD="$(printf '%s' "$INPUT" | python3 -c 'import json,sys
try:
    d=json.load(sys.stdin)
    print(d.get("cwd",""))
except Exception:
    pass' 2>/dev/null)"
fi

[ -z "$COMMAND" ] && exit 0

# Specific attribution phrases, not a bare "claude" match -- a bare match
# would false-positive on a person named Claude, or a commit message that
# just discusses Claude Code as a tool. These are the exact shapes Claude
# Code's own default commit/PR template produces.
ATTRIBUTION_PATTERN='Co-Authored-By:[^"]*[Cc]laude|noreply@anthropic\.com|Generated with \[?Claude|🤖 Generated with|Generated[^"]*Claude Code'

block() {
  echo "[block-ai-attribution] $1" >&2
  echo "Remove the AI/Claude attribution line and try again." >&2
  exit 2
}

contains_attribution() {
  printf '%s' "$1" | grep -qiE "$ATTRIBUTION_PATTERN"
}

# Anchored to the start of the command or right after a shell operator
# (;, &&, ||, |, newline -- newlines are normalized to ';' first so the
# regex itself never has to embed a literal newline, which some grep
# builds choke on), not a bare substring match -- a bare "*git commit*"
# glob would also fire on a command that merely *mentions* that text (e.g.
# an echo, a grep, this very script's own test invocation), which is
# exactly what happened during manual testing of this hook.
is_cmd() {
  printf '%s' "$COMMAND" | tr '\n' ';' | grep -qE "(^|[;&|]) *$1"
}

if is_cmd 'git +commit\b'; then
  if contains_attribution "$COMMAND"; then
    block "This commit message contains AI/Claude attribution."
  fi
elif is_cmd 'gh +(pr|issue) +(create|edit|comment)\b'; then
  if contains_attribution "$COMMAND"; then
    block "This PR/issue text contains AI/Claude attribution."
  fi
elif is_cmd 'git +push\b'; then
  # The push command itself never carries a commit message -- check the
  # actual commits about to go out (e.g. one added via --amend, or made
  # some other way) as a last line of defense. Best-effort: if the
  # outgoing range can't be determined (no upstream yet, cwd not a repo)
  # this silently allows rather than blocking a push it has no way to
  # evaluate.
  range_msgs=""
  if [ -n "$CWD" ] && [ -d "$CWD" ]; then
    range_msgs="$(cd "$CWD" 2>/dev/null && git log --format=%B '@{u}..HEAD' 2>/dev/null)"
  fi
  if [ -z "$range_msgs" ]; then
    range_msgs="$(git log --format=%B '@{u}..HEAD' 2>/dev/null)"
  fi
  if [ -n "$range_msgs" ] && contains_attribution "$range_msgs"; then
    block "One of the commits being pushed contains AI/Claude attribution."
  fi
fi

exit 0
