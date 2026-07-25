# =========================
# -- Outputs
# =========================

output "bootstrap_external_ip" {
  value = google_compute_instance.bootstrap.network_interface[0].access_config[0].nat_ip
}

output "novnc_url" {
  value = "http://${google_compute_instance.bootstrap.network_interface[0].access_config[0].nat_ip}:6080"
}

output "bootstrap_service_account" {
  value = google_service_account.bootstrap.email
}
