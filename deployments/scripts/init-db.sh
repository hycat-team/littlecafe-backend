#!/bin/sh
set -eu

: "${PGHOST:=postgres}"
: "${PGPORT:=5432}"
: "${PGUSER:?PGUSER is required}"
: "${PGPASSWORD:?PGPASSWORD is required}"
: "${PGDATABASE:=postgres}"

export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

require_password() {
  name=$1
  eval "value=\${$name:-}"
  if [ -z "$value" ]; then
    echo "missing required environment variable: $name" >&2
    exit 1
  fi
  printf '%s' "$value"
}

psql_admin() {
  psql -X -q -v ON_ERROR_STOP=1 "$@"
}

create_role_and_database() {
  role=$1
  password=$2
  database=$3

  psql_admin -v role="$role" -v password="$password" -v database="$database" <<'SQL'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'role', :'password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'role')\gexec

SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', :'role', :'password')\gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'database', :'role')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'database')\gexec
SQL
}

services="identity order activity wallet ai notification reporting"

for service in $services; do
  upper=$(printf '%s' "$service" | tr '[:lower:]' '[:upper:]')
  password=$(require_password "${upper}_DB_PASSWORD")
  echo "provisioning database littlecafe_${service}"
  create_role_and_database "littlecafe_${service}" "$password" "littlecafe_${service}"
done

for service in $services; do
  database="littlecafe_${service}"
  echo "ensuring pgcrypto in ${database}"
  psql_admin -d "$database" -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
done

echo "ensuring vector in littlecafe_ai"
psql_admin -d littlecafe_ai -c 'CREATE EXTENSION IF NOT EXISTS vector;'

echo "database bootstrap complete"
