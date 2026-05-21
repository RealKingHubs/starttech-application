#!/bin/bash
set -euo pipefail

BASE_URL="${1:-}"

if [ -z "$BASE_URL" ]; then
  echo "Usage: $0 <base-url>"
  exit 1
fi

URL="${BASE_URL%/}/health"

echo "Checking health endpoint..."

MAX_RETRIES=12
SLEEP_SECONDS=10

for ((i=1; i<=MAX_RETRIES; i++))
do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$URL")

  if [ "$STATUS" -eq 200 ]; then
    echo "Health check passed"
    exit 0
  fi

  echo "Attempt $i failed with status $STATUS"
  echo "Retrying in $SLEEP_SECONDS seconds..."

  sleep $SLEEP_SECONDS
done

echo "Health check failed after retries"
exit 1
