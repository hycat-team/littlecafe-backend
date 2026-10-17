#!/usr/bin/env sh
set -eu

module='github.com/hycat-team/littlecafe-backend'
failed=0

for service_dir in services/*; do
  [ -d "$service_dir" ] || continue
  service=$(basename "$service_dir")
  matches=$(grep -R -n -E --include='*.go' "${module}/services/" "$service_dir" 2>/dev/null || true)
  violations=$(printf '%s\n' "$matches" | grep -v "${module}/services/${service}/" || true)
  if [ -n "$violations" ]; then
    echo "cross-service import detected in $service:" >&2
    echo "$violations" >&2
    failed=1
  fi
done

if grep -R -n -E --include='*.go' "${module}/services/" shared contracts 2>/dev/null; then
  echo "shared/contracts must not depend on a service module" >&2
  failed=1
fi

for domain_dir in shared/domain services/*/internal/domain; do
  [ -d "$domain_dir" ] || continue
  if grep -R -n -E --include='*.go' "${module}/(shared|services/[^/]+)/(application|infrastructure|presentation)/" "$domain_dir" 2>/dev/null; then
    echo "domain layer points outward: $domain_dir" >&2
    failed=1
  fi
done

for application_dir in shared/application services/*/internal/application; do
  [ -d "$application_dir" ] || continue
  if grep -R -n -E --include='*.go' "${module}/(shared|services/[^/]+)/(infrastructure|presentation)/" "$application_dir" 2>/dev/null; then
    echo "application layer points to infrastructure/presentation: $application_dir" >&2
    failed=1
  fi
done

[ "$failed" -eq 0 ]
echo "Dependency boundaries are valid."
