#!/usr/bin/env sh
# Run SOPS through a pinned container image (default) or a local binary.
#
# Usage:
#   sops-run.sh decrypt --input-type dotenv --output-type dotenv file.env.enc
#   SOPS_MODE=binary sops-run.sh edit deployments/env/staging.env.enc
#
# Environment:
#   SOPS_MODE        docker (default) | binary
#   SOPS_IMAGE       pinned image reference (digest), docker mode only
#   SOPS_BINARY      sops executable path, binary mode only (default: sops)
#   SOPS_AGE_KEY_FILE  age identity file; mounted read-only in docker mode
#
# Security notes:
# - The age identity file is mounted, never passed through -e (docker inspect
#   would expose the value).
# - The container runs without network, read-only, no capabilities, and as the
#   calling user so it cannot write to the host as root.
# - `sops edit` needs an editor and is only supported in binary mode.
set -eu

SOPS_MODE=${SOPS_MODE:-docker}
SOPS_IMAGE=${SOPS_IMAGE:-ghcr.io/getsops/sops:v3.13.3-alpine@sha256:ae501277bf742f1662e0f881f43dd8fd6798b489a8058e921dbf6cda597140ea}
SOPS_BINARY=${SOPS_BINARY:-sops}

die() { printf 'sops-run: %s\n' "$*" >&2; exit 1; }

[ "$#" -gt 0 ] || { printf 'usage: sops-run.sh <sops-args>...\n' >&2; exit 64; }

if [ "$SOPS_MODE" = "binary" ]; then
  command -v "$SOPS_BINARY" >/dev/null 2>&1 || die "sops binary not found: $SOPS_BINARY"
  exec "$SOPS_BINARY" "$@"
fi

[ "$SOPS_MODE" = "docker" ] || die "unsupported SOPS_MODE: $SOPS_MODE"
command -v docker >/dev/null 2>&1 || die "docker is required for SOPS_MODE=docker"

case "${1:-}" in
  edit)
    die "sops edit needs an editor; install sops locally and use SOPS_MODE=binary"
    ;;
esac

if ! docker image inspect "$SOPS_IMAGE" >/dev/null 2>&1; then
  docker pull "$SOPS_IMAGE" >/dev/null
fi

user_opt=""
if command -v id >/dev/null 2>&1; then
  user_opt="--user $(id -u):$(id -g)"
fi

key_opt=""
if [ -n "${SOPS_AGE_KEY_FILE:-}" ]; then
  [ -r "$SOPS_AGE_KEY_FILE" ] || die "age key file is not readable: $SOPS_AGE_KEY_FILE"
  key_dir=$(cd "$(dirname "$SOPS_AGE_KEY_FILE")" && pwd)
  key_opt="-v $key_dir:/keys:ro -e SOPS_AGE_KEY_FILE=/keys/$(basename "$SOPS_AGE_KEY_FILE")"
fi

# Word splitting of the optional flags is intentional; the paths involved are
# controlled by the caller and contain no spaces on the VPS.
# shellcheck disable=SC2086
exec docker run --rm \
  --network none \
  --read-only \
  --tmpfs /tmp \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  $user_opt $key_opt \
  -v "$PWD:/repo" -w /repo -e HOME=/tmp \
  "$SOPS_IMAGE" "$@"
