#!/usr/bin/env bash
# sync-env.sh (GCP) — Reads terraform/gcp/persistent outputs into .env.gcp,
# then regenerates terraform/gcp/ephemeral/terraform.tfvars.
# Usage: ./scripts/gcp/sync-env.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
source "$REPO_ROOT/scripts/lib/env.sh"

ENV_FILE="$REPO_ROOT/.env"
CLOUD_ENV_FILE="$REPO_ROOT/.env.gcp"
TF_DIR="$REPO_ROOT/terraform/gcp/persistent"
TFVARS_FILE="$REPO_ROOT/terraform/gcp/ephemeral/terraform.tfvars"

ensure_env_file "$ENV_FILE"
ensure_env_file "$CLOUD_ENV_FILE"

# Read terraform outputs
echo "Reading terraform outputs from $TF_DIR ..."
cd "$TF_DIR"

update_env "$CLOUD_ENV_FILE" "ARTIFACT_BUCKET_NAME" "$(terraform output -raw artifact_bucket_name)"
update_env "$CLOUD_ENV_FILE" "SECRET_STORE_NAME"    "$(terraform output -raw unity_license_secret_id)"

echo ".env.gcp updated with terraform outputs."

# --- Generate terraform.tfvars for the ephemeral config ---

REPO_URL="$(read_env "$ENV_FILE" REPO_URL)"
PLATFORM="$(read_env "$ENV_FILE" PLATFORM)"
ADMIN_PASSWORD="$(read_env "$ENV_FILE" ADMIN_PASSWORD)"
GITHUB_TOKEN="$(read_env "$ENV_FILE" GITHUB_TOKEN)" # optional — empty for public repos
PROJECT_ID="$(read_env "$CLOUD_ENV_FILE" PROJECT_ID)"
ARTIFACT_BUCKET_NAME="$(read_env "$CLOUD_ENV_FILE" ARTIFACT_BUCKET_NAME)"
# Auto-detect the operator's current IP when not pinned manually —
# residential IPs rotate, and the firewall must follow the operator.
NOVNC_ALLOWED_CIDR="$(read_env "$CLOUD_ENV_FILE" NOVNC_ALLOWED_CIDR)"
if [ -z "$NOVNC_ALLOWED_CIDR" ] || [ "$NOVNC_ALLOWED_CIDR" = "auto" ]; then
  NOVNC_ALLOWED_CIDR="$(curl -s https://api.ipify.org)/32"
  echo "Auto-detected operator IP for noVNC firewall: $NOVNC_ALLOWED_CIDR"
fi

if [ -z "$REPO_URL" ] || [ -z "$PLATFORM" ] || [ -z "$ADMIN_PASSWORD" ] || [ -z "$PROJECT_ID" ] || [ -z "$ARTIFACT_BUCKET_NAME" ] || [ -z "$NOVNC_ALLOWED_CIDR" ]; then
  echo "Warning: REPO_URL, PLATFORM, ADMIN_PASSWORD, PROJECT_ID, ARTIFACT_BUCKET_NAME, or NOVNC_ALLOWED_CIDR is empty"
  echo "Skipping terraform.tfvars generation. Fill in .env / .env.gcp and run again."
else
  cat > "$TFVARS_FILE" << EOF
project_id           = "$PROJECT_ID"
repo_url             = "$REPO_URL"
platform             = "$PLATFORM"
admin_password       = "$ADMIN_PASSWORD"
github_token         = "$GITHUB_TOKEN"
artifact_bucket_name = "$ARTIFACT_BUCKET_NAME"
novnc_allowed_cidr   = "$NOVNC_ALLOWED_CIDR"
EOF
  echo "Generated $TFVARS_FILE"
fi

echo "Done."
