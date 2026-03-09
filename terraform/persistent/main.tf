terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location
}

resource "azurerm_key_vault" "main" {
  name                = var.key_vault_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"

  # Admin access for the operator (current az login user)
  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    secret_permissions = ["Get", "Set", "List", "Delete", "Purge"]
  }
}

# --- Shared Image Gallery ---

resource "azurerm_shared_image_gallery" "main" {
  name                = "unity_ci_gallery"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

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

# --- Azure Batch Account ---

resource "azurerm_batch_account" "main" {
  name                                = "unitycibatch"
  location                            = azurerm_resource_group.main.location
  resource_group_name                 = azurerm_resource_group.main.name
  pool_allocation_mode                = "BatchService"
  storage_account_id                  = azurerm_storage_account.function.id
  storage_account_authentication_mode = "StorageKeys"
  public_network_access_enabled       = true
}

# --- Function App ---

resource "azurerm_storage_account" "function" {
  name                     = "unitycifuncstorage"
  location                 = azurerm_resource_group.main.location
  resource_group_name      = azurerm_resource_group.main.name
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_service_plan" "function" {
  name                = "unity-ci-func-plan"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  os_type             = "Linux"
  sku_name            = "Y1" # Consumption Plan
}

resource "azurerm_linux_function_app" "main" {
  name                       = var.function_app_name
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  service_plan_id            = azurerm_service_plan.function.id
  storage_account_name       = azurerm_storage_account.function.name
  storage_account_access_key = azurerm_storage_account.function.primary_access_key

  site_config {}

  app_settings = {
    KEY_VAULT_NAME        = azurerm_key_vault.main.name
    BATCH_ACCOUNT_NAME    = azurerm_batch_account.main.name
    IMAGE_GALLERY_NAME    = azurerm_shared_image_gallery.main.name
    IMAGE_DEFINITION_NAME = azurerm_shared_image.main.name
    RESOURCE_GROUP_NAME   = azurerm_resource_group.main.name
  }
}
