#!/usr/bin/env bash
set -euo pipefail
HOST="${1:-127.0.0.1:8090}"
EMAIL="${2:-admin@example.com}"
PASS="${3:-admin123}"
curl -sS "http://${HOST}/api/admins/auth-with-password" \
  -H 'Content-Type: application/json' \
  -d "{\"identity\":\"${EMAIL}\",\"password\":\"${PASS}\"}" \
  | jq -r '.token'
