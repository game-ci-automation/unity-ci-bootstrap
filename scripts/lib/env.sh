# env.sh — Shared env-file helpers, sourced by scripts/<cloud>/sync-env.sh.
# Env file layout:
#   .env        — common, cloud-agnostic (REPO_URL, PLATFORM, ADMIN_PASSWORD)
#   .env.azure  — Azure resource names
#   .env.gcp    — GCP project/resources

# ensure_env_file <file> — create from its .example if missing
ensure_env_file() {
  if [ ! -f "$1" ]; then
    cp "$1.example" "$1"
    echo "Created $(basename "$1") from example"
  fi
}

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
