#!/usr/bin/env bash
# capture.sh (GCP) — Capture the bootstrap VM as a build-node image, then
# destroy the ephemeral infrastructure. GCP counterpart of scripts/azure/capture.sh.
#
# Run AFTER downloader-gcp finished on the VM (license in Secret Manager,
# game-ci image pre-pulled, Unity Hub removed).
#
# Usage: ./scripts/gcp/capture.sh [zone]   (default: northamerica-northeast2-a)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
ZONE="${1:-northamerica-northeast2-a}"
VM_NAME="unity-bootstrap-vm"
IMAGE_NAME="unity-ci-build-image"
IMAGE_FAMILY="unity-ci-build"

PROJECT_ID="$(grep '^PROJECT_ID=' "$REPO_ROOT/.env.gcp" | cut -d'=' -f2-)"
if [ -z "$PROJECT_ID" ]; then
  echo "PROJECT_ID missing in .env.gcp" >&2
  exit 1
fi

echo "=== Step 1/3: Stop VM (clean disk state for imaging) ==="
gcloud compute instances stop "$VM_NAME" --zone "$ZONE" --project "$PROJECT_ID"

echo ""
echo "=== Step 2/3: Create image from boot disk ==="
# Batch build nodes boot from this image: docker + game-ci editor image
# pre-pulled, license already externalized to Secret Manager.
gcloud compute images create "$IMAGE_NAME" \
  --source-disk "$VM_NAME" \
  --source-disk-zone "$ZONE" \
  --family "$IMAGE_FAMILY" \
  --project "$PROJECT_ID"

echo ""
echo "=== Step 3/3: Destroy ephemeral infrastructure ==="
cd "$REPO_ROOT/terraform/gcp/ephemeral"
terraform destroy -auto-approve

echo ""
echo "=== Capture complete ==="
echo "Image: $IMAGE_NAME (family: $IMAGE_FAMILY, project: $PROJECT_ID)"
