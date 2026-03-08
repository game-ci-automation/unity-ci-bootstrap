variable "location" {
  description = "Azure region"
  type        = string
  default     = "canadacentral"
}

variable "resource_group_name" {
  description = "Name of the Azure resource group"
  type        = string
  default     = "unity-ci-enabler-rg"
}

variable "key_vault_name" {
  description = "Name of the Azure Key Vault"
  type        = string
  default     = "unity-ci-kv"
}

variable "function_app_name" {
  description = "Name of the Azure Function App"
  type        = string
  default     = "unity-ci-func"
}
