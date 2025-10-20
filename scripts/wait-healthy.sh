#!/usr/bin/env bash
set -euo pipefail
HOST="${1:?host:port}"
DEADLINE=$((SECONDS+20))
until curl -sf "http://${HOST}/healthz" >/dev/null; do
  if (( SECONDS > DEADLINE )); then
    echo "timeout waiting for ${HOST}/healthz" >&2; exit 1
  fi
  sleep 0.2
done
