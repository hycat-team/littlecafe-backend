#!/usr/bin/env sh
# Decrypt a SOPS env file into a runtime dotenv file, safely.
#
# Usage:
#   apply-env.sh <encrypted-file> <target-file> [example-file]
#
# - Decrypts to a private temporary file in the target directory.
# - Verifies there are no CR bytes, no leftover SOPS metadata, and that every
#   key documented in the example file is present.
# - Keeps the previous target as <target>.prev for rollback.
# - Replaces the target atomically and sets mode 600.
#
# Never prints secret values.
set -eu

usage() {
  printf 'usage: apply-env.sh <encrypted-file> <target-file> [example-file]\n' >&2
  exit 64
}

fail() { printf 'apply-env: %s\n' "$*" >&2; exit 1; }

[ "$#" -ge 2 ] || usage

encrypted=$1
target=$2
example=${3:-"${target}.example"}

[ -f "$encrypted" ] || fail "encrypted file not found: $encrypted"
[ -f "$example" ] || fail "example file not found: $example"

script_dir=$(cd "$(dirname "$0")" && pwd)
sops_run="${script_dir}/sops-run.sh"
[ -f "$sops_run" ] || fail "missing helper: $sops_run"

umask 077
target_dir=$(cd "$(dirname "$target")" && pwd)
tmp=$(mktemp "${target_dir}/.envtmp.XXXXXX")
keys_tmp=$(mktemp "${target_dir}/.envkeys.XXXXXX")
cleanup() { rm -f "$tmp" "$keys_tmp"; }
trap cleanup EXIT

SOPS_MODE=${SOPS_MODE:-docker} sh "$sops_run" \
  decrypt --input-type dotenv --output-type dotenv "$encrypted" >"$tmp"

[ -s "$tmp" ] || fail "decrypted file is empty"

if LC_ALL=C grep -q "$(printf '\r')" "$tmp"; then
  fail "decrypted file contains CR bytes (normalize the source to LF)"
fi

if grep -q '^sops_' "$tmp"; then
  fail "decrypted file still contains SOPS metadata keys"
fi

grep -aoE '^[A-Za-z_][A-Za-z0-9_]*=' "$example" | sed 's/=$//' | sort -u >"$keys_tmp"

missing=""
while IFS= read -r key; do
  [ -n "$key" ] || continue
  if ! grep -q "^${key}=" "$tmp"; then
    missing="${missing} ${key}"
  fi
done <"$keys_tmp"

[ -z "$missing" ] || fail "missing keys from ${example}:${missing}"

key_count=$(grep -cE '^[A-Za-z_][A-Za-z0-9_]*=' "$tmp" || true)

if [ -f "$target" ]; then
  cp -p "$target" "${target}.prev"
  chmod 600 "${target}.prev" 2>/dev/null || true
fi

chmod 600 "$tmp"
mv "$tmp" "$target"
rm -f "$keys_tmp"
trap - EXIT

printf 'apply-env: wrote %s (%s keys, mode 600)\n' "$target" "$key_count"
