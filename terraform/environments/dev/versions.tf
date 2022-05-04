terraform {
  backend "gcs" {
    bucket  = "weave-gitops-enterprise-terraform-state"
    prefix  = "dev"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.19.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.12.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.4.0"
    }
  }

  required_version = ">= 1.1.9"
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

provider "azurerm" {
  features {}

  subscription_id = "6bf943cd-75d6-4e1c-b2bf-b8691841d4ae"
}