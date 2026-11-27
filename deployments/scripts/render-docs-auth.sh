#!/usr/bin/env sh
# Render the nginx basic-auth file for staging API docs from the decrypted env.
#
# Usage: render-docs-auth.sh <env-file> <output-file>
#
# Reads DOCS_BASIC_AUTH_USER / DOCS_BASIC_AUTH_PASSWORD from the dotenv file,
# hashes the password with openssl (apr1), and writes <user>:<hash>.
# Never prints the password or hash.
set -eu

usage() {
  printf 'usage: render-docs-auth.sh <env-file> <output-file>\n' >&2
  exit 64
}

fail() { printf 'render-docs-auth: %s\n' "$*" >&2; exit 1; }

[ "$#" -eq 2 ] || usage
env_file=$1
out_file=$2

[ -f "$env_file" ] || fail "env file not found: $env_file"
command -v openssl >/dev/null 2>&1 || fail "openssl is required"

read_key() {
  sed -n "s/^$1=//p" "$env_file" | tail -n 1 | tr -d '\r'
}

user=$(read_key DOCS_BASIC_AUTH_USER)
pass=$(read_key DOCS_BASIC_AUTH_PASSWORD)
[ -n "$user" ] || fail "DOCS_BASIC_AUTH_USER is empty"
[ -n "$pass" ] || fail "DOCS_BASIC_AUTH_PASSWORD is empty"

hash=$(printf '%s' "$pass" | openssl passwd -apr1 -stdin)
dest_dir=$(dirname "$out_file")
mkdir -p "$dest_dir"
printf '%s:%s\n' "$user" "$hash" > "$out_file"
chmod 755 "$dest_dir"
chmod 644 "$out_file"
printf 'render-docs-auth: wrote %s\n' "$out_file"
