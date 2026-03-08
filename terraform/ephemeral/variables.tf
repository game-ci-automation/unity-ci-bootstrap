variable "location" {
  description = "Azure region"
  type        = string
  default     = "canadacentral"
}

variable "resource_group_name" {
  description = "Name of the existing persistent resource group"
  type        = string
  default     = "unity-ci-enabler-rg"
}

variable "key_vault_name" {
  description = "Name of the existing persistent Key Vault"
  type        = string
  default     = "unity-ci-kv"
}

variable "vm_size" {
  description = "Azure VM size"
  type        = string
  default     = "Standard_D4s_v3"
}

variable "admin_username" {
  description = "Admin username for the VM"
  type        = string
  default     = "azureuser"
}

variable "admin_password" {
  description = "Admin password for the VM"
  type        = string
  sensitive   = true
}

variable "function_app_url" {
  description = "URL of the Azure Function App (from persistent output)"
  type        = string
}

variable "repo_url" {
  description = "GitHub repository URL (e.g. https://github.com/user/repo)"
  type        = string
}
