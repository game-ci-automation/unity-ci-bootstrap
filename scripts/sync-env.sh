#!/usr/bin/env bash
# sync-env.sh — Reads terraform persistent outputs and writes them into the
# cloud-specific env file (.env.azure / .env.gcp), then regenerates the
# ephemeral terraform.tfvars for that cloud.
#
# Usage: ./scripts/sync-env.sh [azure|gcp]   (default: azure)
#
# Env file layout:
#   .env        — common, cloud-agnostic (REPO_URL, PLATFORM, GITHUB_TOKEN, ADMIN_PASSWORD)
#   .env.azure  — Azure resource names   (auto-filled here)
#   .env.gcp    — GCP project/resources  (manual IDs + auto-filled outputs)

set -euo pipefail

CLOUD="${1:-azure}"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$REPO_ROOT/.env"
CLOUD_ENV_FILE="$REPO_ROOT/.env.$CLOUD"
TF_DIR="$REPO_ROOT/terraform/$CLOUD/persistent"

# Create env files from examples if they don't exist
for f in "$ENV_FILE" "$CLOUD_ENV_FILE"; do
  if [ ! -f "$f" ]; then
    cp "$f.example" "$f"
    echo "Created $(basename "$f") from example"
  fi
done

# --- Helpers ---

# update_env <file> <key> <value>
update_env() {
  local file="$1" key="$2" value="$3"
  if grep -q "^${key}=" "$file"; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$file"
  else
    echo "${key}=${value}" >> "$file"
  fi
}

# read_env <file> <key>
read_env() {
  grep "^${2}=" "$1" | cut -d'=' -f2-
}

# Read terraform outputs
echo "Reading terraform outputs from $TF_DIR ..."
cd "$TF_DIR"

case "$CLOUD" in
  azure)
    update_env "$CLOUD_ENV_FILE" "RESOURCE_GROUP_NAME"   "$(terraform output -raw resource_group_name)"
    update_env "$CLOUD_ENV_FILE" "KEY_VAULT_NAME"        "$(terraform output -raw key_vault_name)"
    update_env "$CLOUD_ENV_FILE" "FUNCTION_APP_URL"      "$(terraform output -raw function_app_url)"
    update_env "$CLOUD_ENV_FILE" "FUNCTION_APP_NAME"     "$(terraform output -raw function_app_name)"
    update_env "$CLOUD_ENV_FILE" "BATCH_ACCOUNT_NAME"    "$(terraform output -raw batch_account_name)"
    update_env "$CLOUD_ENV_FILE" "IMAGE_GALLERY_NAME"    "$(terraform output -raw image_gallery_name)"
    update_env "$CLOUD_ENV_FILE" "IMAGE_DEFINITION_NAME" "$(terraform output -raw image_definition_name)"
    ;;
  gcp)
    update_env "$CLOUD_ENV_FILE" "ARTIFACT_BUCKET_NAME" "$(terraform output -raw artifact_bucket_name)"
    update_env "$CLOUD_ENV_FILE" "SECRET_STORE_NAME"    "$(terraform output -raw unity_license_secret_id)"
    ;;
  *)
    echo "Unknown cloud: $CLOUD (expected azure or gcp)" >&2
    exit 1
    ;;
esac

echo "$(basename "$CLOUD_ENV_FILE") updated with terraform outputs."

# --- Generate terraform.tfvars for the ephemeral config ---

ADMIN_PASSWORD="$(read_env "$ENV_FILE" ADMIN_PASSWORD)"
REPO_URL="$(read_env "$ENV_FILE" REPO_URL)"
PLATFORM="$(read_env "$ENV_FILE" PLATFORM)"

if [ "$CLOUD" = "azure" ]; then
  TFVARS_FILE="$REPO_ROOT/terraform/azure/ephemeral/terraform.tfvars"

  # GITHUB_TOKEN is Azure-only (legacy webhook registration flow)
  GITHUB_TOKEN="$(read_env "$CLOUD_ENV_FILE" GITHUB_TOKEN)"
  FUNCTION_APP_URL="$(read_env "$CLOUD_ENV_FILE" FUNCTION_APP_URL)"
  KEY_VAULT_NAME="$(read_env "$CLOUD_ENV_FILE" KEY_VAULT_NAME)"
  RESOURCE_GROUP_NAME="$(read_env "$CLOUD_ENV_FILE" RESOURCE_GROUP_NAME)"
  IMAGE_GALLERY_NAME="$(read_env "$CLOUD_ENV_FILE" IMAGE_GALLERY_NAME)"
  IMAGE_DEFINITION_NAME="$(read_env "$CLOUD_ENV_FILE" IMAGE_DEFINITION_NAME)"

  if [ -z "$ADMIN_PASSWORD" ] || [ -z "$REPO_URL" ] || [ -z "$FUNCTION_APP_URL" ] || [ -z "$PLATFORM" ]; then
    echo "Warning: ADMIN_PASSWORD, REPO_URL, FUNCTION_APP_URL, or PLATFORM is empty"
    echo "Skipping terraform.tfvars generation. Fill in .env / .env.azure and run again."
  else
    cat > "$TFVARS_FILE" << EOF
admin_password        = "$ADMIN_PASSWORD"
repo_url              = "$REPO_URL"
function_app_url      = "$FUNCTION_APP_URL"
platform              = "$PLATFORM"
key_vault_name        = "$KEY_VAULT_NAME"
resource_group_name   = "$RESOURCE_GROUP_NAME"
image_gallery_name    = "$IMAGE_GALLERY_NAME"
image_definition_name = "$IMAGE_DEFINITION_NAME"
github_token          = "$GITHUB_TOKEN"
EOF
    echo "Generated $TFVARS_FILE"
  fi
else
  # GCP ephemeral tfvars generation is added with terraform/gcp/ephemeral (Issue C)
  echo "GCP ephemeral tfvars generation: not yet implemented (Issue C)."
fi

echo "Done."
