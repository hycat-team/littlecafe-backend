#!/usr/bin/env sh
# Compare the local .env file with .env.example.
#
# Missing keys are an error; extra local-only keys are reported as warnings.
set -eu

[ -f .env.example ] || { echo "env-local-check: .env.example not found (run from repo root)" >&2; exit 1; }
[ -f .env ] || { echo "env-local-check: .env not found; run 'make env-local'" >&2; exit 1; }

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

grep -oE '^[A-Za-z_][A-Za-z0-9_]*=' .env.example | sed 's/=$//' | sort -u >"$tmp/example"
grep -oE '^[A-Za-z_][A-Za-z0-9_]*=' .env | sed 's/=$//' | sort -u >"$tmp/local"

missing=$(comm -23 "$tmp/example" "$tmp/local")
if [ -n "$missing" ]; then
  echo "env-local-check: .env is missing keys:" >&2
  printf '%s\n' "$missing" >&2
  exit 1
fi

extra=$(comm -13 "$tmp/example" "$tmp/local")
if [ -n "$extra" ]; then
  echo "env-local-check: warning: local-only keys (not in .env.example):"
  printf '%s\n' "$extra"
fi

echo "env-local-check: .env matches .env.example ($(wc -l <"$tmp/local" | tr -d ' ') keys)"
