#!/usr/bin/env sh
# Example: the telemetry sink swallows any report and returns 200 {} so the
# console's play-report upload succeeds without anything reaching Nintendo.
set -e
BASE="${1:-http://localhost:8472}"
echo "== POST a fake play report =="
curl -s -X POST "$BASE/v1/播report" -H 'Content-Type: application/octet-stream' \
  --data-binary 'nextendo test telemetry' ; echo
echo "(any method/path returns 200 {}; set TELEMETRY_DUMP to log bodies)"
