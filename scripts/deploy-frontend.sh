#!/bin/bash
set -euo pipefail

BUCKET_NAME="${1:-}"
SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPTS_DIR/.." && pwd)"

if [ -z "$BUCKET_NAME" ]; then
  echo "Usage: $0 <frontend-bucket-name>"
  exit 1
fi

aws s3 sync "$REPO_ROOT/frontend/dist/" "s3://$BUCKET_NAME" --delete
