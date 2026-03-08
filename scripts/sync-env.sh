#!/usr/bin/env bash
# sync-env.sh — Reads terraform persistent outputs and writes them into .env
# Usage: ./scripts/sync-env.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$REPO_ROOT/.env"
TF_DIR="$REPO_ROOT/terraform/persistent"

# Create .env from .env.example if it doesn't exist
if [ ! -f "$ENV_FILE" ]; then
  cp "$REPO_ROOT/.env.example" "$ENV_FILE"
  echo "Created .env from .env.example"
fi

# Read terraform outputs
echo "Reading terraform outputs from $TF_DIR ..."
cd "$TF_DIR"

update_env() {
  local key="$1"
  local value="$2"
  if grep -q "^${key}=" "$ENV_FILE"; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$ENV_FILE"
  else
    echo "${key}=${value}" >> "$ENV_FILE"
  fi
}

update_env "RESOURCE_GROUP_NAME" "$(terraform output -raw resource_group_name)"
update_env "KEY_VAULT_NAME"      "$(terraform output -raw key_vault_name)"
update_env "FUNCTION_APP_URL"    "$(terraform output -raw function_app_url)"
update_env "FUNCTION_APP_NAME"   "$(terraform output -raw function_app_name)"
update_env "BATCH_ACCOUNT_NAME"  "$(terraform output -raw batch_account_name)"
update_env "IMAGE_GALLERY_NAME"  "$(terraform output -raw image_gallery_name)"
update_env "IMAGE_DEFINITION_NAME" "$(terraform output -raw image_definition_name)"

echo ".env updated with terraform outputs."

# --- Generate terraform.tfvars for ephemeral ---
TFVARS_FILE="$REPO_ROOT/terraform/ephemeral/terraform.tfvars"

read_env() {
  grep "^${1}=" "$ENV_FILE" | cut -d'=' -f2-
}

ADMIN_PASSWORD="$(read_env ADMIN_PASSWORD)"
REPO_URL="$(read_env REPO_URL)"
FUNCTION_APP_URL="$(read_env FUNCTION_APP_URL)"

if [ -z "$ADMIN_PASSWORD" ] || [ -z "$REPO_URL" ] || [ -z "$FUNCTION_APP_URL" ]; then
  echo "Warning: ADMIN_PASSWORD, REPO_URL, or FUNCTION_APP_URL is empty in .env"
  echo "Skipping terraform.tfvars generation. Fill in .env and run again."
else
  cat > "$TFVARS_FILE" << EOF
admin_password   = "$ADMIN_PASSWORD"
repo_url         = "$REPO_URL"
function_app_url = "$FUNCTION_APP_URL"
EOF
  echo "Generated $TFVARS_FILE"
fi

echo "Done."
