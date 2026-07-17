# =========================
# -- Terraform Configuration
# =========================
# Declares which provider plugin to download.
# `terraform init` fetches the plugin (like `npm install`).

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"   # HashiCorp's official Azure plugin
      version = "~> 3.0"              # Any 3.x version, but not 4.0 (breaking change guard)
    }
  }
}

# =========================
# -- Provider Initialization
# =========================
# Activates the azurerm plugin. Uses credentials from `az login`.
# `features {}` is a required empty block — azurerm won't start without it.

provider "azurerm" {
  features {}
}

# =========================
# -- Current User Info (read-only)
# =========================
# `data` = read-only query. Doesn't create anything in Azure.
# Fetches: "who am I?" from the current `az login` session.
#   - data.azurerm_client_config.current.tenant_id   → Azure AD tenant
#   - data.azurerm_client_config.current.object_id   → my user ID
# Used below for Key Vault admin access policy.

data "azurerm_client_config" "current" {}

# =========================
# -- Resource Group
# =========================
# Container for all Azure resources. Everything belongs to one resource group.
# "main" is a local name (used in code to reference: azurerm_resource_group.main.name).
# var.resource_group_name comes from variables.tf (default: "unity-ci-enabler-rg").

resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location
}

# =========================
# -- Key Vault
# =========================
# Secure secret storage. Stores:
#   - WEBHOOK-SECRET (GitHub webhook HMAC signing key)
#   - UNITY-LICENSE  (Unity license file, uploaded by bootstrap VM)
#
# access_policy (inline) = admin permissions for the operator (current az login user).
# Function App and VM get their own separate access policies below.
#
# lifecycle { ignore_changes = [access_policy] }:
#   Terraform treats inline access_policy as "the complete list".
#   Without this, every `terraform apply` would DELETE policies added
#   by separate azurerm_key_vault_access_policy resources (e.g. function_app).
#   This tells Terraform: "don't touch access_policy after initial creation."

resource "azurerm_key_vault" "main" {
  name                = var.key_vault_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"

  # Use Azure RBAC for authorization instead of access_policy.
  # This eliminates the inline access_policy vs separate resource conflict,
  # and the lifecycle { ignore_changes } hack is no longer needed.
  enable_rbac_authorization = true
}

# =========================
# -- Shared Image Gallery
# =========================
# Stores the specialized VM image created during bootstrap.
# Gallery = container, Image Definition = metadata, Image Version = actual image.
# Image Version (1.0.0) is created by capture.sh (not Terraform),
# which is why `terraform destroy` can't delete it automatically.

resource "azurerm_shared_image_gallery" "main" {
  name                = "unity_ci_gallery"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

# Image Definition: describes what kind of images this gallery holds.
# specialized = true: preserves machine ID → Unity license stays valid.
# hyper_v_generation = "V2": must match the VM that creates the image (immutable after creation).

resource "azurerm_shared_image" "main" {
  name                = "unity-ci-build-image"
  gallery_name        = azurerm_shared_image_gallery.main.name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  os_type             = "Linux"
  hyper_v_generation  = "V2"
  specialized         = true

  identifier {
    publisher = "UnityCIEnabler"
    offer     = "BuildNode"
    sku       = "specialized"
  }
}

# =========================
# -- Azure Batch Account
# =========================
# Runs build jobs. Function App submits jobs here via REST API.
# pool_allocation_mode = "BatchService": Azure manages pool infrastructure.
# storage_account_authentication_mode = "StorageKeys": required by azurerm provider.

resource "azurerm_batch_account" "main" {
  name                                = "unitycibatch"
  location                            = azurerm_resource_group.main.location
  resource_group_name                 = azurerm_resource_group.main.name
  pool_allocation_mode                = "BatchService"
  storage_account_id                  = azurerm_storage_account.function.id
  storage_account_authentication_mode = "StorageKeys"
  public_network_access_enabled       = true
}

# =========================
# -- Storage Account (for Function App)
# =========================
# Function App needs a storage account for its runtime (zip deployment, logs).
# Also linked to Batch Account above.
# LRS = Locally Redundant Storage (cheapest, single datacenter).

resource "azurerm_storage_account" "function" {
  name                     = "unitycifuncstorage"
  location                 = azurerm_resource_group.main.location
  resource_group_name      = azurerm_resource_group.main.name
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

# =========================
# -- Function App (Service Plan + App)
# =========================
# Hosts the Go webhook handler (unity-ci-function repo).
# Y1 = Consumption Plan: pay per execution, $0 when idle.
# Code is NOT deployed here — deploy.sh does that separately.

resource "azurerm_service_plan" "function" {
  name                = "unity-ci-func-plan"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  os_type             = "Linux"
  sku_name            = "Y1"   # Consumption Plan — pay per execution, free tier available
}

# The actual Function App. Created empty — unity-ci-function repo's deploy.sh
# uploads the Go binary later.
#
# identity { type = "SystemAssigned" }:
#   Azure automatically creates a Managed Identity for this app.
#   The Go code uses DefaultAzureCredential() to get tokens — zero credentials in code.
#
# app_settings: environment variables the Go binary reads at startup.
#   KEY_VAULT_NAME   → to read WEBHOOK-SECRET
#   BATCH_ACCOUNT_URL → to submit Batch jobs
#   IMAGE_RESOURCE_ID → to reference the VM image in autoPool
#
# Note: VM_SIZE is NOT here. deploy.sh sets it via `az functionapp config appsettings set`
# because Terraform would delete it on every apply (app_settings is treated as complete list).

resource "azurerm_linux_function_app" "main" {
  name                       = var.function_app_name
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  service_plan_id            = azurerm_service_plan.function.id
  storage_account_name       = azurerm_storage_account.function.name
  storage_account_access_key = azurerm_storage_account.function.primary_access_key

  site_config {}

  identity {
    type = "SystemAssigned"
  }

  app_settings = {
    KEY_VAULT_NAME    = azurerm_key_vault.main.name
    BATCH_ACCOUNT_URL = "https://${azurerm_batch_account.main.name}.${azurerm_resource_group.main.location}.batch.azure.com"
    IMAGE_RESOURCE_ID = azurerm_shared_image.main.id
  }
}

# =========================
# -- RBAC: Function App Permissions (Least Privilege)
# =========================
# Each permission was discovered through runtime errors, not pre-granted:
#   1. 403 PermissionDenied → added Batch Contributor
#   2. 502 BadGateway       → fixed Key Vault policy conflict (policy since replaced by RBAC)
#   3. 400 InsufficientPermissions → added Gallery Reader
#   4. 202 success
#
# Function App's Managed Identity (principal_id) gets the minimum permissions needed.

# Batch Account: Contributor — allows creating Jobs and autoPools
resource "azurerm_role_assignment" "function_batch_contributor" {
  scope                = azurerm_batch_account.main.id
  role_definition_name = "Contributor"
  principal_id         = azurerm_linux_function_app.main.identity[0].principal_id
}

# Image Gallery: Reader — autoPool needs to reference the VM image
resource "azurerm_role_assignment" "function_gallery_reader" {
  scope                = azurerm_shared_image_gallery.main.id
  role_definition_name = "Reader"
  principal_id         = azurerm_linux_function_app.main.identity[0].principal_id
}

# Key Vault: Secret read only — reads WEBHOOK-SECRET at startup.
# Uses RBAC (Key Vault Secrets User) instead of access_policy.
# Cannot write, delete, or manage secrets.
resource "azurerm_role_assignment" "function_keyvault_reader" {
  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = azurerm_linux_function_app.main.identity[0].principal_id
}

# Operator: full secret management (create, read, update, delete secrets)
resource "azurerm_role_assignment" "operator_keyvault_admin" {
  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Secrets Officer"
  principal_id         = data.azurerm_client_config.current.object_id
}
