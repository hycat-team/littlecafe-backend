#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
REGISTRY_FILE=${ERROR_CODES_FILE:-"$REPO_ROOT/../littlecafe-docs/api/error-codes.md"}
OPENAPI_DIR="$REPO_ROOT/contracts/openapi"

if [ ! -f "$REGISTRY_FILE" ]; then
    echo "error-codes-check: registry not found at $REGISTRY_FILE; skipping because littlecafe-docs is a separate repository."
    exit 0
fi

code_pattern='(IDENTITY|ORDER|ACTIVITY|WALLET|AI|NOTIFICATION|REPORTING)_[0-9]{4}'

openapi_codes=$(find "$OPENAPI_DIR" -type f \( -name '*.yaml' -o -name '*.yml' \) -exec grep -hoE "$code_pattern" {} + 2>/dev/null | sort -u || true)
registry_codes=$(awk -F'|' '/^\| *[A-Z]+_[0-9]{4} *\|/ {gsub(/ /, "", $2); print $2}' "$REGISTRY_FILE" | sort -u)

failed=0

for code in $openapi_codes; do
    if ! grep -Eq "^\|[[:space:]]*$code[[:space:]]*\|" "$REGISTRY_FILE"; then
        echo "error-codes-check: $code is used in OpenAPI but missing from $REGISTRY_FILE" >&2
        failed=1
    fi
done

for code in $registry_codes; do
    if ! printf '%s\n' "$openapi_codes" | grep -qx "$code"; then
        echo "error-codes-check: $code is registered but not used by any OpenAPI contract" >&2
        failed=1
    fi
done

if [ "$failed" -ne 0 ]; then
    exit 1
fi

if [ -z "$openapi_codes" ]; then
    echo "error-codes-check: no business error codes declared in OpenAPI."
else
    echo "error-codes-check: OpenAPI and registry agree."
fi
