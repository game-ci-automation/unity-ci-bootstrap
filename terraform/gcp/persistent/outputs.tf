# =========================
# -- Outputs
# =========================
# Cross-cloud contract: same *meanings* as terraform/azure/persistent outputs,
# consumed later by sync-env / ephemeral tfvars generation.
#   Azure key_vault_name  ↔ GCP secret_store (secret id)
#   Azure storage account ↔ GCP artifact_bucket_name
#   Azure resource group  ↔ GCP project_id (the container concept)

output "project_id" {
  value = google_project.main.project_id
}

output "region" {
  value = var.region
}

output "artifact_bucket_name" {
  value = google_storage_bucket.artifacts.name
}

output "unity_license_secret_id" {
  value = google_secret_manager_secret.unity_license.secret_id
}
