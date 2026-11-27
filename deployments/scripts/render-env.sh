#!/usr/bin/env sh
# Render a runtime dotenv file from a template and a decrypted dotenv source.
#
# Usage: render-env.sh <example-file> <values-file> <output-file>
#
# - Key order, section comments, and blank lines come from the example file.
# - Values come from the decrypted file (first '=' separates key and value).
# - Keys present in the values file but absent from the example are appended
#   under an "# Extra keys" section.
# - Never prints values.
set -eu

usage() {
  printf 'usage: render-env.sh <example-file> <values-file> <output-file>\n' >&2
  exit 64
}

fail() { printf 'render-env: %s\n' "$*" >&2; exit 1; }

[ "$#" -eq 3 ] || usage
example=$1
values=$2
out=$3

[ -f "$example" ] || fail "example file not found: $example"
[ -f "$values" ] || fail "values file not found: $values"

tmp="${out}.tmp.$$"
trap 'rm -f "$tmp"' EXIT

VALUES_FILE=$values awk '
  function fail(msg) {
    printf "render-env: %s\n", msg > "/dev/stderr"
    exit 1
  }
  BEGIN {
    values = ENVIRON["VALUES_FILE"]
    while ((getline line < values) > 0) {
      sub(/\r$/, "", line)
      if (line ~ /^[A-Za-z_][A-Za-z0-9_]*=/) {
        eq = index(line, "=")
        key = substr(line, 1, eq - 1)
        val[key] = substr(line, eq + 1)
      }
    }
    close(values)
  }
  {
    line = $0
    sub(/\r$/, "", line)
    if (line ~ /^[A-Za-z_][A-Za-z0-9_]*=/) {
      eq = index(line, "=")
      key = substr(line, 1, eq - 1)
      if (!(key in val)) fail("missing value for " key)
      print key "=" val[key]
      seen[key] = 1
    } else {
      print line
    }
  }
  END {
    n = 0
    for (key in val) {
      if (!(key in seen)) {
        extras[++n] = key
      }
    }
    for (i = 2; i <= n; i++) {
      current = extras[i]
      j = i - 1
      while (j > 0 && extras[j] > current) {
        extras[j + 1] = extras[j]
        j--
      }
      extras[j + 1] = current
    }
    if (n > 0) {
      print ""
      print "# Extra keys not present in the example"
      for (i = 1; i <= n; i++) {
        print extras[i] "=" val[extras[i]]
      }
    }
  }
' "$example" >"$tmp"
mv "$tmp" "$out"
trap - EXIT
