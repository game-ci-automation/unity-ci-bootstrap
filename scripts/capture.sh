#!/usr/bin/env bash
# capture.sh — Captures VM image to gallery, deletes VM, and destroys ephemeral infra.
# Run from repo root after downloader completes on the VM.
# Usage: ./scripts/capture.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$REPO_ROOT/.env"

if [ ! -f "$ENV_FILE" ]; then
  echo "Error: .env not found. Run scripts/sync-env.sh first."
  exit 1
fi

read_env() {
  grep "^${1}=" "$ENV_FILE" | cut -d'=' -f2-
}

RG="$(read_env RESOURCE_GROUP_NAME)"
GALLERY="$(read_env IMAGE_GALLERY_NAME)"
IMAGE_DEF="$(read_env IMAGE_DEFINITION_NAME)"
VM_NAME="unity-ci-vm"

if [ -z "$RG" ] || [ -z "$GALLERY" ] || [ -z "$IMAGE_DEF" ]; then
  echo "Error: RESOURCE_GROUP_NAME, IMAGE_GALLERY_NAME, or IMAGE_DEFINITION_NAME is empty in .env"
  exit 1
fi

echo "=== Step 1/4: Deallocate VM ==="
az vm deallocate --resource-group "$RG" --name "$VM_NAME"
echo "VM deallocated."

echo ""
echo "=== Step 2/4: Capture VM image ==="
VMID=$(az vm show -g "$RG" -n "$VM_NAME" --query id -o tsv)
az sig image-version create \
  --resource-group "$RG" \
  --gallery-name "$GALLERY" \
  --gallery-image-definition "$IMAGE_DEF" \
  --gallery-image-version 1.0.0 \
  --virtual-machine "$VMID"
echo "Image captured to gallery."

echo ""
echo "=== Step 3/4: Delete VM ==="
az vm delete --resource-group "$RG" --name "$VM_NAME" --yes
echo "VM deleted."

echo ""
echo "=== Step 4/4: Destroy ephemeral infrastructure ==="
cd "$REPO_ROOT/terraform/ephemeral"
terraform destroy -auto-approve -var="github_token=dummy"
echo "Ephemeral resources destroyed."

echo ""
echo "=== Capture complete ==="
echo "Image: $GALLERY/$IMAGE_DEF/1.0.0"
