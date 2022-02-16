terraform {
  backend "gcs" {
    bucket  = "weave-gitops-enterprise-terraform-state"
    prefix  = "dev"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.10.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0.0"
    }
  }

  required_version = ">= 1.1.5"
}

variable "project" {
  description = "GCP project id"
}

variable "region" {
  description = "GCP project region"
}

provider "google" {
  project = var.project
  region = var.region
}