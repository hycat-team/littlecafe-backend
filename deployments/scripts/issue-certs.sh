#!/bin/sh
set -eu

LEGO_IMAGE=${LEGO_IMAGE:-goacme/lego:v5.5.2}
LEGO_DATA=${LEGO_DATA:-/opt/littlecafe/secrets/acme}
LETSENCRYPT_DIR=${LETSENCRYPT_DIR:-/etc/letsencrypt}
VERCEL_TOKEN_FILE=${VERCEL_TOKEN_FILE:-/opt/littlecafe/secrets/vercel.token}
ACME_EMAIL=${ACME_EMAIL:-}
ACME_EMAIL_FILE=${ACME_EMAIL_FILE:-/opt/littlecafe/secrets/acme/account.email}
CERT_DOMAINS=${CERT_DOMAINS:-"api.littlecafe.kids staging-api.littlecafe.kids"}

if [ -z "$ACME_EMAIL" ] && [ -f "$ACME_EMAIL_FILE" ]; then
  ACME_EMAIL=$(cat "$ACME_EMAIL_FILE")
fi

if [ -z "$ACME_EMAIL" ]; then
  echo "ACME_EMAIL is required (Let's Encrypt account email)" >&2
  exit 1
fi

if [ ! -f "$VERCEL_TOKEN_FILE" ]; then
  echo "Vercel token file not found: $VERCEL_TOKEN_FILE" >&2
  exit 1
fi

VERCEL_API_TOKEN=$(cat "$VERCEL_TOKEN_FILE")
export VERCEL_API_TOKEN
if [ -n "${VERCEL_TEAM_ID:-}" ]; then
  export VERCEL_TEAM_ID
fi

mkdir -p "$LEGO_DATA" "$LETSENCRYPT_DIR/live"

run_lego() {
  docker run --rm \
    --user "$(id -u):$(id -g)" \
    -v "${LEGO_DATA}:/data" \
    -e VERCEL_API_TOKEN \
    ${VERCEL_TEAM_ID:+-e VERCEL_TEAM_ID} \
    "${LEGO_IMAGE}" \
    --email "$ACME_EMAIL" --dns vercel --accept-tos --path /data "$@"
}

install_cert() {
  domain=$1
  cert_dir="$LETSENCRYPT_DIR/live/$domain"
  crt="$LEGO_DATA/certificates/$domain.crt"
  key="$LEGO_DATA/certificates/$domain.key"
  issuer="$LEGO_DATA/certificates/$domain.issuer.crt"

  if [ ! -f "$crt" ] || [ ! -f "$key" ]; then
    echo "missing lego output for $domain" >&2
    return 1
  fi

  mkdir -p "$cert_dir"
  if [ -f "$issuer" ]; then
    cat "$crt" "$issuer" > "$cert_dir/fullchain.pem"
  else
    cp "$crt" "$cert_dir/fullchain.pem"
  fi
  cp "$key" "$cert_dir/privkey.pem"
  chmod 644 "$cert_dir/fullchain.pem"
  chmod 600 "$cert_dir/privkey.pem"
  echo "installed certificate for $domain"
}

issue_missing() {
  for domain in $CERT_DOMAINS; do
    if [ ! -f "$LEGO_DATA/certificates/$domain.crt" ]; then
      echo "issuing certificate for $domain"
      run_lego --domains "$domain" run
    fi
  done
}

renew_all() {
  echo "renewing existing certificates"
  run_lego renew --days 30
}

command=${1:-issue}

case "$command" in
  issue)
    has_existing=0
    for domain in $CERT_DOMAINS; do
      [ -f "$LEGO_DATA/certificates/$domain.crt" ] && has_existing=1
    done
    [ "$has_existing" -eq 1 ] && renew_all
    issue_missing
    for domain in $CERT_DOMAINS; do
      install_cert "$domain"
    done
    ;;
  renew)
    renew_all
    for domain in $CERT_DOMAINS; do
      install_cert "$domain"
    done
    ;;
  *)
    echo "usage: $0 [issue|renew]" >&2
    exit 64
    ;;
esac

echo "certificate operation complete"
