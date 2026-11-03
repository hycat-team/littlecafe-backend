#!/usr/bin/env sh
# Create or inspect the age identity used by SOPS on this machine.
#
# Usage:
#   sh scripts/env/init-dev-key.sh        # create if missing, then print the public key
#   sh scripts/env/init-dev-key.sh --new  # add another identity to the same file
#
# Only the public key (age1...) is safe to share. Never send the private key.
set -eu

key_file="${XDG_CONFIG_HOME:-$HOME/.config}/sops/age/keys.txt"

if ! command -v age-keygen >/dev/null 2>&1; then
  printf 'age-keygen not found. Install age first (apt install age, brew install age, ...)\n' >&2
  exit 1
fi

mkdir -p "$(dirname "$key_file")"

if [ -f "$key_file" ] && [ "${1:-}" != "--new" ]; then
  pub=$(sed -n 's/^# public key: //p' "$key_file" | tail -n1)
  if [ -n "$pub" ]; then
    printf 'Existing age identity found.\n'
    printf 'Public key (safe to share): %s\n' "$pub"
    printf 'Private key file (never share or commit): %s\n' "$key_file"
    printf 'Run with --new to add another identity to this file.\n'
    exit 0
  fi
fi

output=$(age-keygen 2>/dev/null)
printf '%s\n' "$output" >>"$key_file"
chmod 600 "$key_file" 2>/dev/null || true
pub=$(printf '%s\n' "$output" | sed -n 's/^# public key: //p')

printf 'Created a new age identity.\n'
printf 'Public key (safe to share): %s\n' "$pub"
printf 'Private key file (never share or commit): %s\n' "$key_file"
printf 'Back up this file in your password manager; losing it only requires re-enrolment.\n'
