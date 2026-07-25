# =========================
# -- Terraform Configuration
# =========================
# GCP counterpart of terraform/azure/ephemeral: a temporary desktop VM for
# Unity license activation. Destroyed after the license is in Secret Manager
# and the VM is captured as a build-node image (scripts/gcp/capture.sh).

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
}

provider "google" {
  project = var.project_id
  zone    = var.zone
}

# =========================
# -- Service Account (zero credential)
# =========================
# Attached to the VM; downloader-gcp authenticates via ADC through the
# metadata server — no key files anywhere (Azure Managed Identity parity).

resource "google_service_account" "bootstrap" {
  account_id   = "unity-bootstrap-vm"
  display_name = "Unity CI bootstrap VM"
}

# =========================
# -- Least-Privilege IAM (resource-level, not project-level)
# =========================
# secretVersionAdder on the one secret: can add license versions,
# cannot read them back. Narrower than the Azure ephemeral VM's
# Secrets Officer role.

resource "google_secret_manager_secret_iam_member" "license_adder" {
  project   = var.project_id
  secret_id = "unity-license"
  role      = "roles/secretmanager.secretVersionAdder"
  member    = "serviceAccount:${google_service_account.bootstrap.email}"
}

# objectViewer on the artifacts bucket: cloud-init fetches the
# downloader-gcp binary from bootstrap/ (until a GitHub release ships
# per-cloud asset names).

resource "google_storage_bucket_iam_member" "artifact_reader" {
  bucket = var.artifact_bucket_name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.bootstrap.email}"
}

# =========================
# -- Firewall: noVNC
# =========================
# Default network already allows ssh (22); add browser desktop access (6080).
# x11vnc runs -nopw, so reachability IS the auth boundary — restrict the
# source to the operator's IP (novnc_allowed_cidr in terraform.tfvars).

resource "google_compute_firewall" "novnc" {
  name    = "unity-bootstrap-allow-novnc"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["6080"]
  }

  source_ranges = [var.novnc_allowed_cidr]
  target_tags   = ["unity-bootstrap"]
}

# =========================
# -- Bootstrap VM
# =========================
# Ubuntu on GCE runs cloud-init from the `user-data` metadata key —
# same cloud-init mechanism as Azure custom_data.

resource "google_compute_instance" "bootstrap" {
  name         = "unity-bootstrap-vm"
  machine_type = var.machine_type
  zone         = var.zone
  tags         = ["unity-bootstrap"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 128
      type  = "pd-ssd"
    }
  }

  network_interface {
    network = "default"
    access_config {} # ephemeral external IP (noVNC + ssh)
  }

  service_account {
    email  = google_service_account.bootstrap.email
    scopes = ["cloud-platform"] # actual access is bounded by IAM above
  }

  # Shared cloud-init template; GCP-specific bits injected as variables.
  metadata = {
    user-data = templatefile("${path.module}/../../../cloud-init/bootstrap.yaml", {
      vm_user      = "gcpuser"
      create_user  = true  # no osProfile equivalent — cloud-init creates the user
      install_gh   = false # Discord-triggered design: no cloud↔GitHub link
      vnc_password = var.admin_password
      env_exports = <<-EOT
        export REPO_URL=${var.repo_url}
        export PLATFORM=${var.platform}
        export PROJECT_ID=${var.project_id}
        export GITHUB_TOKEN=${var.github_token}
      EOT
      # From the artifacts bucket until a GitHub release ships per-cloud
      # assets; VM service account has objectViewer on the bucket.
      downloader_fetch = "gsutil cp \"gs://${var.artifact_bucket_name}/bootstrap/downloader-gcp-linux-amd64\" /home/gcpuser/Desktop/downloader || (DEBIAN_FRONTEND=noninteractive apt-get install -y google-cloud-cli && gsutil cp \"gs://${var.artifact_bucket_name}/bootstrap/downloader-gcp-linux-amd64\" /home/gcpuser/Desktop/downloader)"
      public_ip_probe  = "curl -s -H \"Metadata-Flavor:Google\" \"http://169.254.169.254/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip\""
    })
  }
}
