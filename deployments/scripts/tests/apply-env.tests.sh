#!/usr/bin/env bash
# End-to-end test for apply-env.sh using a throwaway age identity and a
# synthetic fixture. The fixture contains no real secret.
#
# Requires: sops (or Docker for SOPS_MODE=docker) and age-keygen.
set -eu

root=$(cd "$(dirname "$0")/../../.." && pwd)
cd "$root"

if ! command -v age-keygen >/dev/null 2>&1; then
  printf 'SKIP: age-keygen not installed\n'
  exit 0
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

age-keygen -o "$tmp/keys.txt" >/dev/null 2>&1
pub=$(sed -n 's/^# public key: //p' "$tmp/keys.txt" | head -n1)
[ -n "$pub" ] || { printf 'FAIL: could not read test public key\n' >&2; exit 1; }

cat >"$tmp/fixture.env" <<'EOF'
# fixture comment
APP_MODE=all
EMPTY_VALUE=
WITH_SPACES=hello world
WITH_EQUALS=left=right
WITH_QUOTES=quote"inside
EOF

cp "$tmp/fixture.env" "$tmp/example.env"

SOPS_MODE=${SOPS_MODE:-binary}
export SOPS_MODE
export SOPS_AGE_KEY_FILE="$tmp/keys.txt"

# Encrypt from the scratch directory so the repository .sops.yaml is not
# picked up (the fixture must use the throwaway identity, not real recipients).
(
  cd "$tmp"
  sh "$root/deployments/scripts/sops-run.sh" encrypt \
    --input-type dotenv --output-type dotenv --age "$pub" fixture.env \
    >fixture.env.enc
)

# Seed the target so the previous value must be preserved for rollback.
printf 'OLD_VALUE=1\n' >"$tmp/target.env"

sh deployments/scripts/apply-env.sh \
  "$tmp/fixture.env.enc" "$tmp/target.env" "$tmp/example.env" >/dev/null

# Blank lines are dropped by the SOPS dotenv store; compare meaningful lines.
if ! diff -B "$tmp/fixture.env" "$tmp/target.env" >"$tmp/diff.txt"; then
  printf 'FAIL: round-trip mismatch\n' >&2
  cat "$tmp/diff.txt" >&2
  exit 1
fi

if [ ! -f "$tmp/target.env.prev" ]; then
  printf 'FAIL: expected a .prev backup\n' >&2
  exit 1
fi

if ls "$tmp"/.envkeys.* >/dev/null 2>&1; then
  printf 'FAIL: apply-env left temporary key files behind\n' >&2
  exit 1
fi

# A wrong identity must fail, not silently produce an empty file.
if SOPS_AGE_KEY_FILE="$tmp/other-keys.txt" sh deployments/scripts/apply-env.sh \
  "$tmp/fixture.env.enc" "$tmp/target2.env" "$tmp/example.env" >/dev/null 2>&1; then
  printf 'FAIL: decryption with a missing identity should fail\n' >&2
  exit 1
fi

# The example controls key order/comments; keys absent from the example are
# appended under an extra-keys section.
cat >"$tmp/source2.env" <<'EOF'
C=3
A=left=right
B=hello world
D=extra
EOF

cat >"$tmp/example2.env" <<'EOF'
# section one
A=placeholder
B=placeholder
# section two
C=placeholder
EOF

(
  cd "$tmp"
  sh "$root/deployments/scripts/sops-run.sh" encrypt \
    --input-type dotenv --output-type dotenv --age "$pub" source2.env \
    >source2.env.enc
)

sh deployments/scripts/apply-env.sh \
  "$tmp/source2.env.enc" "$tmp/target3.env" "$tmp/example2.env" >/dev/null

cat >"$tmp/expected2.env" <<'EOF'
# section one
A=left=right
B=hello world
# section two
C=3

# Extra keys not present in the example
D=extra
EOF

if ! diff -u "$tmp/expected2.env" "$tmp/target3.env" >"$tmp/diff2.txt"; then
  printf 'FAIL: rendered order/comments mismatch\n' >&2
  cat "$tmp/diff2.txt" >&2
  exit 1
fi

# The renderer must fail when the example declares a key without a value.
cat >"$tmp/example3.env" <<'EOF'
A=placeholder
MISSING=placeholder
EOF
if sh deployments/scripts/render-env.sh \
  "$tmp/example3.env" "$tmp/source2.env" "$tmp/target4.env" >/dev/null 2>&1; then
  printf 'FAIL: render-env should fail on a missing value\n' >&2
  exit 1
fi

printf 'PASS: apply-env round-trip and failure paths\n'
