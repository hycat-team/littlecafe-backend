#!/usr/bin/env sh
# Validate encrypted env files without decrypting them.
#
# Checks:
# - every key documented in .env.<env>.example exists in the encrypted file;
# - no unexpected extra keys (legacy keys are allowlisted until pruned);
# - no CR bytes;
# - SOPS metadata is present.
#
# Run from the repository root. No age key is required.
set -eu

ENV_DIR=${ENV_DIR:-deployments/env}
ENVS=${ENVS:-"production staging"}

# Deprecated per-service database keys still present in the live VPS files.
# They are unused by compose/init-db and will be pruned after cutover.
LEGACY_ALLOWLIST="IDENTITY_DB_PASSWORD ORDER_DB_PASSWORD ACTIVITY_DB_PASSWORD WALLET_DB_PASSWORD NOTIFICATION_DB_PASSWORD REPORTING_DB_PASSWORD"

fail() { printf 'env-check: %s\n' "$*" >&2; exit 1; }

extract_keys() {
  grep -aoE '^[A-Za-z_][A-Za-z0-9_]*=' "$1" | sed 's/=$//' | grep -v '^sops_' | sort -u
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

status=0
for env in $ENVS; do
  enc="${ENV_DIR}/${env}.env.enc"
  example=".env.${env}.example"

  [ -f "$enc" ] || fail "missing encrypted file: $enc"
  [ -f "$example" ] || fail "missing example file: $example"

  if LC_ALL=C grep -q "$(printf '\r')" "$enc"; then
    fail "$enc contains CR bytes; commit with LF endings"
  fi

  if ! grep -q '^sops_mac=' "$enc"; then
    fail "$enc has no SOPS metadata (not encrypted?)"
  fi

  extract_keys "$enc" >"$tmp/enc.${env}"
  extract_keys "$example" >"$tmp/example.${env}"

  missing=$(comm -23 "$tmp/example.${env}" "$tmp/enc.${env}")
  if [ -n "$missing" ]; then
    printf 'env-check: %s is missing keys:\n%s\n' "$enc" "$missing" >&2
    status=1
  fi

  extra=$(comm -13 "$tmp/example.${env}" "$tmp/enc.${env}")
  for key in $extra; do
    case " $LEGACY_ALLOWLIST " in
      *" $key "*) ;;
      *)
        printf 'env-check: %s has unexpected key: %s\n' "$enc" "$key" >&2
        status=1
        ;;
    esac
  done

  printf 'env-check: %s OK (%s keys)\n' "$enc" "$(grep -c . "$tmp/enc.${env}" || true)"
done

[ "$status" -eq 0 ] || exit 1
