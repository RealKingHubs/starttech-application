#!/bin/bash
set -euo pipefail

IMAGE_INPUT="${1:-}"
ACCOUNT_ID="093796422475"
REGION="us-east-1"
REPOSITORY="dev-starttech-backend"
SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -z "$IMAGE_INPUT" ]; then
  echo "Usage: $0 <image-uri-or-tag>"
  echo "Examples:"
  echo "  $0 123456789012.dkr.ecr.us-east-1.amazonaws.com/dev-starttech-backend:abc123"
  echo "  $0 abc123"
  exit 1
fi

if [[ "$IMAGE_INPUT" == *"/"*":"* ]]; then
  IMAGE_URI="$IMAGE_INPUT"
else
  IMAGE_URI="$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com/$REPOSITORY:$IMAGE_INPUT"
fi

echo "Rolling back backend to image: $IMAGE_URI"
"$SCRIPTS_DIR/deploy-backend.sh" "$IMAGE_URI"
echo "Rollback command sent successfully."
