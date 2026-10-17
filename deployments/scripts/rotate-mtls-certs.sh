#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  echo "usage: $0 /absolute/path/to/certs" >&2
  exit 64
fi

output_dir=$1
case "$output_dir" in
  /*) ;;
  *) echo "certificate directory must be an absolute path" >&2; exit 64 ;;
esac

command -v openssl >/dev/null 2>&1 || {
  echo "openssl is required" >&2
  exit 69
}

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
case "$output_dir/" in
  "$repo_root"/*) echo "certificate directory must be outside the repository" >&2; exit 64 ;;
esac

parent_dir=$(dirname -- "$output_dir")
mkdir -p -- "$parent_dir"
stage_dir=$(mktemp -d "$parent_dir/.littlecafe-certs.XXXXXX")
trap 'rm -rf -- "$stage_dir"' EXIT HUP INT TERM

ca_days=${LITTLECAFE_CA_DAYS:-3650}
leaf_days=${LITTLECAFE_CERT_DAYS:-90}

openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "$stage_dir/ca.key"
openssl req -x509 -new -sha256 -key "$stage_dir/ca.key" -days "$ca_days" \
  -subj "/O=LittleCafe/CN=LittleCafe Private CA" -out "$stage_dir/ca.crt"

issue_certificate() {
  name=$1
  usage=$2
  san=$3
  openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 -out "$stage_dir/$name.key"
  openssl req -new -sha256 -key "$stage_dir/$name.key" -subj "/O=LittleCafe/CN=$name" -out "$stage_dir/$name.csr"
  {
    echo "basicConstraints=critical,CA:FALSE"
    echo "keyUsage=critical,digitalSignature,keyAgreement"
    echo "extendedKeyUsage=$usage"
    echo "subjectAltName=$san"
  } > "$stage_dir/$name.ext"
  openssl x509 -req -sha256 -in "$stage_dir/$name.csr" -CA "$stage_dir/ca.crt" \
    -CAkey "$stage_dir/ca.key" -CAcreateserial -days "$leaf_days" \
    -extfile "$stage_dir/$name.ext" -out "$stage_dir/$name.crt"
  rm -f -- "$stage_dir/$name.csr" "$stage_dir/$name.ext"
}

for service in identity order activity wallet; do
  issue_certificate "$service" serverAuth "DNS:$service,DNS:$service.littlecafe.internal"
done
issue_certificate service-client clientAuth "DNS:service-client.littlecafe.internal"

chmod 700 "$stage_dir"
chmod 600 "$stage_dir"/*.key
chmod 644 "$stage_dir"/*.crt

if [ -e "$output_dir" ]; then
  backup_dir="$output_dir.backup.$(date -u +%Y%m%dT%H%M%SZ)"
  mv -- "$output_dir" "$backup_dir"
  echo "previous certificate set moved to $backup_dir"
fi
mv -- "$stage_dir" "$output_dir"
trap - EXIT HUP INT TERM
echo "mTLS certificate set written to $output_dir"
