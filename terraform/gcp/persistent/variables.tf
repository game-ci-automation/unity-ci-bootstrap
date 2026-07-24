# =========================
# -- Variables
# =========================
# Per-user values (org, billing, project id) have NO defaults on purpose —
# they are personal identifiers, supplied via terraform.tfvars (gitignored),
# which is generated from .env (see .env.example, GCP section).

variable "org_id" {
  description = "GCP organization ID this project is created under"
  type        = string
}

variable "billing_account" {
  description = "Billing account ID the project is linked to"
  type        = string
}

variable "project_id" {
  description = "Globally-unique GCP project ID (also prefixes the artifact bucket name)"
  type        = string
}

variable "region" {
  description = "Default region for regional resources (artifact bucket)"
  type        = string
  default     = "northamerica-northeast2" # Toronto (GCP's canada-central equivalent)
}
