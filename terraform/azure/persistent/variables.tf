# =========================
# -- Input Variables
# =========================
# These are the inputs to this Terraform module.
# Referenced in main.tf as: var.location, var.resource_group_name, etc.
# Can be overridden via:
#   - terraform.tfvars file
#   - CLI: terraform apply -var="location=eastus"
#   - Environment: TF_VAR_location=eastus

variable "location" {
  description = "Azure region where all resources are created"
  type        = string
  default     = "canadacentral"
}

variable "resource_group_name" {
  description = "Name of the Azure resource group (container for all resources)"
  type        = string
  default     = "unity-ci-enabler-rg"
}

variable "key_vault_name" {
  description = "Name of the Azure Key Vault (stores WEBHOOK-SECRET + UNITY-LICENSE)"
  type        = string
  default     = "unity-ci-kv"
}

variable "function_app_name" {
  description = "Name of the Azure Function App (hosts the Go webhook handler)"
  type        = string
  default     = "unity-ci-func"
}
