output "resource_group_name" {
  value = azurerm_resource_group.main.name
}

output "key_vault_name" {
  value = azurerm_key_vault.main.name
}

output "key_vault_id" {
  value = azurerm_key_vault.main.id
}

output "function_app_url" {
  value = "https://${azurerm_linux_function_app.main.default_hostname}"
}

output "function_app_name" {
  value = azurerm_linux_function_app.main.name
}

output "image_gallery_name" {
  value = azurerm_shared_image_gallery.main.name
}

output "image_definition_name" {
  value = azurerm_shared_image.main.name
}

output "batch_account_name" {
  value = azurerm_batch_account.main.name
}
