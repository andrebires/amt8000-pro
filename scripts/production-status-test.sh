#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${AMT_HOST:-}" ]]; then
  echo "AMT_HOST is required" >&2
  exit 2
fi

if [[ -z "${AMT_PASSWORD:-}" ]]; then
  echo "AMT_PASSWORD is required" >&2
  exit 2
fi

export AMT_PORT="${AMT_PORT:-9009}"
export AMT_HTTP_ADDR="${AMT_HTTP_ADDR:-127.0.0.1:18080}"

report_dir="docs/test-runs"
mkdir -p "$report_dir"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
report="$report_dir/${stamp}-status.md"

go run ./cmd/amt8000-pro >"/tmp/amt8000-pro-${stamp}.log" 2>&1 &
pid="$!"
trap 'kill "$pid" >/dev/null 2>&1 || true' EXIT

cookie_jar="/tmp/amt8000-pro-${stamp}.cookies"
for _ in $(seq 1 40); do
  if curl -fsS "http://${AMT_HTTP_ADDR}/login" >/dev/null; then
    break
  fi
  sleep 0.25
done

curl -fsS \
  -c "$cookie_jar" \
  -d "host=${AMT_HOST}" \
  -d "port=${AMT_PORT}" \
  -d "password=${AMT_PASSWORD}" \
  "http://${AMT_HTTP_ADDR}/login" >/tmp/amt8000-pro-login.html

curl -fsS \
  -b "$cookie_jar" \
  "http://${AMT_HTTP_ADDR}/api/status" >/tmp/amt8000-pro-status.json

if [[ ! -s /tmp/amt8000-pro-status.json ]]; then
  echo "status request failed" >&2
  cat "/tmp/amt8000-pro-${stamp}.log" >&2 || true
  exit 1
fi

{
  echo "# Production Status Test"
  echo
  echo "- Timestamp UTC: ${stamp}"
  echo "- Panel host: ${AMT_HOST}"
  echo "- Panel port: ${AMT_PORT}"
  echo "- Result: pass"
  echo
  echo "## Response"
  echo
  echo '```json'
  cat /tmp/amt8000-pro-status.json
  echo
  echo '```'
} >"$report"

echo "Wrote $report"
