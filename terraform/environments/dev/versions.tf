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
      version = "~> 4.8.0"
    }
  }

  required_version = ">= 1.1.5"
}

variable "project" {
  description = "GCP project id"
}

variable "gcp_region" {
  description = "GCP project region"
}

variable "aws_region" {
  description = "AWS region"
}

provider "google" {
  project = var.project
  region = var.gcp_region
}

provider "google-beta" {
  project = var.project
  region = var.gcp_region
}

provider "aws" {  
  region = var.aws_region
}