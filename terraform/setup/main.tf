variable "project" {
  description = "GCP project id"
}

variable "location" {
  description = "GCP project location"
}

provider "google" {
  project = var.project
}

resource "google_storage_bucket" "terraform_state" {
  name     = "weave-gitops-enterprise-terraform-state"
  location = var.location
}

