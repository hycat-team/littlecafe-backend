#!/usr/bin/env bash
# Deploy one LittleCafe release from the VPS.
#
# Usage:
#   deploy-release.sh <compose-file> <env-file> <deploy-env-file> \
#     <service=image-ref> [<service=image-ref> ...]
#
# The script writes the selected image refs into the deploy env file atomically,
# pulls images, boots infrastructure, runs database bootstrap, runs migrations in
# parallel, rolls out the selected services, and recreates Nginx last.
set -Eeuo pipefail

usage() {
  cat >&2 <<'EOF'
usage: deploy-release.sh <compose-file> <env-file> <deploy-env-file> <service=image-ref> [...]
EOF
  exit 64
}

log() { printf '==> %s\n' "$*"; }
fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

[ "$#" -ge 4 ] || usage

compose_file=$1
env_file=$2
deploy_file=$3
shift 3

[ -f "$compose_file" ] || fail "compose file not found: $compose_file"
[ -f "$env_file" ] || fail "env file not found: $env_file"

services=()
declare -A images=()
for pair in "$@"; do
  service=${pair%%=*}
  image=${pair#*=}
  case " commerce experience ai platform " in
    *" $service "*) ;;
    *) fail "unsupported service: $service" ;;
  esac
  [ -n "$image" ] || fail "missing image ref for $service"
  services+=("$service")
  images["$service"]=$image
done
[ "${#services[@]}" -gt 0 ] || usage

# One deploy at a time per deploy env file.
if command -v flock >/dev/null 2>&1; then
  exec 9>"${deploy_file}.lock"
  flock -n 9 || fail "another deploy is already running for ${deploy_file}"
else
  printf 'WARN: flock not available; skipping deploy lock\n' >&2
fi

backup="${deploy_file}.prev"
if [ -f "$deploy_file" ]; then
  cp -a "$deploy_file" "$backup"
else
  : > "$deploy_file"
fi
# Keep exactly one rollback copy (.prev); drop legacy per-deploy snapshots.
rm -f "${deploy_file}".bak.*
chmod 600 "$deploy_file" 2>/dev/null || true
chmod 600 "$backup" 2>/dev/null || true

tmp=$(mktemp "${deploy_file}.XXXXXX")
cleanup() { rm -f "$tmp" "${tmp}.new"; }
trap cleanup EXIT

# Merge image refs atomically. One writer, no read-modify-write race.
cp -a "$deploy_file" "$tmp"
for service in "${services[@]}"; do
  upper=$(printf '%s' "$service" | tr '[:lower:]' '[:upper:]')
  key="${upper}_IMAGE"
  value=${images[$service]}
  awk -v key="$key" -v value="$value" '
    BEGIN { done = 0 }
    index($0, key "=") == 1 { print key "=" value; done = 1; next }
    { print }
    END { if (!done) print key "=" value }
  ' "$tmp" > "${tmp}.new"
  mv "${tmp}.new" "$tmp"
done
mv "$tmp" "$deploy_file"
chmod 600 "$deploy_file"
trap - EXIT

dc=(docker compose --env-file "$env_file" --env-file "$deploy_file" -f "$compose_file")

app_state=0 # 0=old release, 1=rolling out new, 2=new release active
restore_env() {
  if [ -f "$backup" ]; then
    cp -a "$backup" "$deploy_file"
    chmod 600 "$deploy_file"
  fi
}

rollback() {
  printf 'ROLLBACK: restoring previous image refs from %s\n' "$backup" >&2
  restore_env
  "${dc[@]}" up -d --wait --wait-timeout 240 --no-deps "${services[@]}" || true
  "${dc[@]}" up -d --force-recreate --no-deps nginx || true
}

on_error() {
  local code=$?
  printf 'deploy failed (exit %s) at line %s\n' "$code" "${1:-?}" >&2
  if [ "$app_state" -eq 1 ]; then
    rollback
  elif [ "$app_state" -eq 0 ]; then
    restore_env
  fi
  exit "$code"
}
trap 'on_error $LINENO' ERR

log "validating compose"
"${dc[@]}" config >/dev/null

log "pulling images: ${services[*]}"
"${dc[@]}" pull "${services[@]}"

log "ensuring infrastructure"
"${dc[@]}" up -d --wait --wait-timeout 180 postgres redis rabbitmq

log "running database bootstrap"
"${dc[@]}" --profile init run --rm db-init

log "running migrations in parallel: ${services[*]}"
declare -A pids=()
for service in "${services[@]}"; do
  migration_log="/tmp/littlecafe-migrate-${service}.log"
  "${dc[@]}" --profile migrate run --rm "migrate-${service}" >"$migration_log" 2>&1 &
  pids["$service"]=$!
done

migration_failed=0
for service in "${services[@]}"; do
  if ! wait "${pids[$service]}"; then
    printf 'migration failed: %s\n' "$service" >&2
    tail -n 80 "/tmp/littlecafe-migrate-${service}.log" >&2 || true
    migration_failed=1
  fi
done
[ "$migration_failed" -eq 0 ] || fail "aborting before application rollout"

log "rolling out services: ${services[*]}"
app_state=1
"${dc[@]}" up -d --wait --wait-timeout 240 --no-deps "${services[@]}"
app_state=2

log "recreating nginx"
"${dc[@]}" up -d --force-recreate --no-deps nginx

log "deployment status"
"${dc[@]}" ps

log "image refs"
grep -E '^(COMMERCE|EXPERIENCE|AI|PLATFORM)_IMAGE=' "$deploy_file" || true

log "deployment complete"
