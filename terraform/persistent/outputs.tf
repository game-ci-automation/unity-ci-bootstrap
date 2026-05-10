# =========================
# -- Outputs
# =========================
# Values exported after `terraform apply`.
# Used by:
#   - sync-env.sh: reads these via `terraform output -raw <name>` → writes to .env
#   - .env → terraform/ephemeral/terraform.tfvars (ephemeral VM needs these values)
#
# These are NOT secrets — they're resource names/URLs that other components need.

output "resource_group_name" {
  value = azurerm_resource_group.main.name
}

output "key_vault_name" {
  value = azurerm_key_vault.main.name
}

output "key_vault_id" {
  value = azurerm_key_vault.main.id
}

# Full URL for the Function App (e.g. https://unity-ci-func.azurewebsites.net)
# Used by: downloader → registers GitHub webhook pointing to this URL
output "function_app_url" {
  value = "https://${azurerm_linux_function_app.main.default_hostname}"
}

output "function_app_name" {
  value = azurerm_linux_function_app.main.name
}

# Gallery/Image names — used by ephemeral VM's cloud-init
# so the downloader knows where to reference the image
output "image_gallery_name" {
  value = azurerm_shared_image_gallery.main.name
}

output "image_definition_name" {
  value = azurerm_shared_image.main.name
}

output "batch_account_name" {
  value = azurerm_batch_account.main.name
}
