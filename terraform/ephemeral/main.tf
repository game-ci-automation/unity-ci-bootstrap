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

data "azurerm_resource_group" "main" {
  name = var.resource_group_name
}

data "azurerm_key_vault" "main" {
  name                = var.key_vault_name
  resource_group_name = var.resource_group_name
}

# Virtual Network
resource "azurerm_virtual_network" "main" {
  name                = "unity-ci-vnet"
  location            = data.azurerm_resource_group.main.location
  resource_group_name = data.azurerm_resource_group.main.name
  address_space       = ["10.0.0.0/16"]
}

# Subnet
resource "azurerm_subnet" "main" {
  name                 = "unity-ci-subnet"
  resource_group_name  = data.azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.1.0/24"]
}

# Public IP
resource "azurerm_public_ip" "main" {
  name                = "unity-ci-pip"
  location            = data.azurerm_resource_group.main.location
  resource_group_name = data.azurerm_resource_group.main.name
  allocation_method   = "Static"
}

# Network Security Group
resource "azurerm_network_security_group" "main" {
  name                = "unity-ci-nsg"
  location            = data.azurerm_resource_group.main.location
  resource_group_name = data.azurerm_resource_group.main.name

  security_rule {
    name                       = "allow-novnc"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "6080"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "allow-ssh"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

# Network Interface
resource "azurerm_network_interface" "main" {
  name                = "unity-ci-nic"
  location            = data.azurerm_resource_group.main.location
  resource_group_name = data.azurerm_resource_group.main.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.main.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.main.id
  }
}

# Attach NSG to NIC
resource "azurerm_network_interface_security_group_association" "main" {
  network_interface_id      = azurerm_network_interface.main.id
  network_security_group_id = azurerm_network_security_group.main.id
}

# Linux Virtual Machine
resource "azurerm_linux_virtual_machine" "main" {
  name                = "unity-ci-vm"
  location            = data.azurerm_resource_group.main.location
  resource_group_name = data.azurerm_resource_group.main.name
  size                = var.vm_size
  admin_username      = var.admin_username
  admin_password      = var.admin_password

  disable_password_authentication = false

  network_interface_ids = [azurerm_network_interface.main.id]

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
    disk_size_gb         = 128
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  custom_data = base64encode(templatefile("${path.module}/../../cloud-init/cloud-init.yaml", {
    function_app_url      = var.function_app_url
    repo_url              = var.repo_url
    platform              = var.platform
    key_vault_name        = var.key_vault_name
    resource_group_name   = var.resource_group_name
    image_gallery_name    = var.image_gallery_name
    image_definition_name = var.image_definition_name
    github_token          = var.github_token
  }))

  identity {
    type = "SystemAssigned"
  }
}

# =========================
# -- Cloud-init Progress Monitor
# =========================
# After VM creation, SSH in and stream cloud-init logs in real time.
# terraform apply will show [STEP 1/10] ~ [STEP 10/10] in the terminal.
# Exits automatically when "BOOTSTRAP VM SETUP COMPLETE" appears.

resource "null_resource" "cloud_init_monitor" {
  depends_on = [azurerm_linux_virtual_machine.main]

  connection {
    type     = "ssh"
    host     = azurerm_public_ip.main.ip_address
    user     = var.admin_username
    password = var.admin_password
  }

  provisioner "remote-exec" {
    inline = [
      "tail -f /var/log/cloud-init-output.log | sed '/SETUP COMPLETE/q'",
      "echo ''",
      "echo '========================================'",
      "echo '  noVNC ready: http://${azurerm_public_ip.main.ip_address}:6080'",
      "echo '========================================'"
    ]
  }
}

# Key Vault RBAC for VM managed identity.
# VM needs to read + write secrets (upload UNITY-LICENSE and WEBHOOK-SECRET).
# Uses Key Vault Secrets Officer (read + write + delete) via RBAC.
# This identity is auto-destroyed when capture.sh deletes the VM.
resource "azurerm_role_assignment" "vm_keyvault" {
  scope                = data.azurerm_key_vault.main.id
  role_definition_name = "Key Vault Secrets Officer"
  principal_id         = azurerm_linux_virtual_machine.main.identity[0].principal_id
}
