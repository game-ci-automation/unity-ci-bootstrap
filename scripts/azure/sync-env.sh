#!/usr/bin/env bash
# sync-env.sh (Azure) — Reads terraform/azure/persistent outputs into
# .env.azure, then regenerates terraform/azure/ephemeral/terraform.tfvars.
# Usage: ./scripts/azure/sync-env.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
source "$REPO_ROOT/scripts/lib/env.sh"

ENV_FILE="$REPO_ROOT/.env"
CLOUD_ENV_FILE="$REPO_ROOT/.env.azure"
TF_DIR="$REPO_ROOT/terraform/azure/persistent"
TFVARS_FILE="$REPO_ROOT/terraform/azure/ephemeral/terraform.tfvars"

ensure_env_file "$ENV_FILE"
ensure_env_file "$CLOUD_ENV_FILE"

# Read terraform outputs
echo "Reading terraform outputs from $TF_DIR ..."
cd "$TF_DIR"

update_env "$CLOUD_ENV_FILE" "RESOURCE_GROUP_NAME"   "$(terraform output -raw resource_group_name)"
update_env "$CLOUD_ENV_FILE" "KEY_VAULT_NAME"        "$(terraform output -raw key_vault_name)"
update_env "$CLOUD_ENV_FILE" "FUNCTION_APP_URL"      "$(terraform output -raw function_app_url)"
update_env "$CLOUD_ENV_FILE" "FUNCTION_APP_NAME"     "$(terraform output -raw function_app_name)"
update_env "$CLOUD_ENV_FILE" "BATCH_ACCOUNT_NAME"    "$(terraform output -raw batch_account_name)"
update_env "$CLOUD_ENV_FILE" "IMAGE_GALLERY_NAME"    "$(terraform output -raw image_gallery_name)"
update_env "$CLOUD_ENV_FILE" "IMAGE_DEFINITION_NAME" "$(terraform output -raw image_definition_name)"

echo ".env.azure updated with terraform outputs."

# --- Generate terraform.tfvars for the ephemeral config ---

ADMIN_PASSWORD="$(read_env "$ENV_FILE" ADMIN_PASSWORD)"
REPO_URL="$(read_env "$ENV_FILE" REPO_URL)"
PLATFORM="$(read_env "$ENV_FILE" PLATFORM)"
GITHUB_TOKEN="$(read_env "$ENV_FILE" GITHUB_TOKEN)"

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

echo "Done."
