# =========================
# -- Terraform Configuration
# =========================
# GCP counterpart of terraform/azure/persistent.
# Scoped to the license bootstrap cycle (serverless CI design, Discord-triggered):
# no GitHub-webhook resources, no function resources yet.

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

# Uses Application Default Credentials (`gcloud auth application-default login`).
# No project set at provider level — the project itself is created below.
provider "google" {}

# =========================
# -- Project
# =========================
# GCP has no Resource Group; the project is the container for everything.
# Created under the jindokim-kor-org organization (not "No organization")
# and linked to the billing account so resources can be provisioned.
# deletion_policy = "DELETE": `terraform destroy` shuts the project down
# (default "PREVENT" would refuse).

resource "google_project" "main" {
  name            = "Unity CI GCP"
  project_id      = var.project_id
  org_id          = var.org_id
  billing_account = var.billing_account
  deletion_policy = "DELETE"
}

# =========================
# -- APIs
# =========================
# GCP requires each service API to be enabled per-project before use
# (Azure has no equivalent step — services are always callable).
# disable_on_destroy = false: destroying this resource just abandons the
# enablement instead of force-disabling the API, which can wedge teardown.

resource "google_project_service" "apis" {
  for_each = toset([
    "compute.googleapis.com",       # bootstrap VM (ephemeral) + Batch node VMs
    "batch.googleapis.com",         # build execution (Azure Batch equivalent)
    "secretmanager.googleapis.com", # license storage (Key Vault equivalent)
    "storage.googleapis.com",       # artifact bucket (Blob Storage equivalent)
  ])
  project            = google_project.main.project_id
  service            = each.value
  disable_on_destroy = false
}

# =========================
# -- Artifact Bucket
# =========================
# Build artifacts (Unity build output) land here.
# Bucket names are globally unique, so the project id is used as prefix.
# force_destroy = true: allow `terraform destroy` even when builds exist.

resource "google_storage_bucket" "artifacts" {
  project                     = google_project.main.project_id
  name                        = "${var.project_id}-artifacts"
  location                    = var.region
  uniform_bucket_level_access = true
  force_destroy               = true

  depends_on = [google_project_service.apis]
}

# =========================
# -- Unity License Secret
# =========================
# Container only — the actual license version (.ulf content) is added by
# downloader-gcp running on the bootstrap VM, mirroring how the Azure
# downloader uploads UNITY-LICENSE to Key Vault.
# No webhook secret here: GitHub connects only to Discord in this design.

resource "google_secret_manager_secret" "unity_license" {
  project   = google_project.main.project_id
  secret_id = "unity-license"

  replication {
    auto {}
  }

  depends_on = [google_project_service.apis]
}
