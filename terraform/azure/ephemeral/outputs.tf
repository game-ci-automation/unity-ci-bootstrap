output "public_ip_address" {
  description = "Public IP address of the VM (for noVNC browser access)"
  value       = azurerm_public_ip.main.ip_address
}

output "vm_id" {
  description = "Resource ID of the VM"
  value       = azurerm_linux_virtual_machine.main.id
}

output "novnc_url" {
  description = "noVNC browser access URL (auto-connect, auto-resize)"
  value       = "http://${azurerm_public_ip.main.ip_address}:6080/vnc.html?resize=scale&clip=true&autoconnect=true"
}
