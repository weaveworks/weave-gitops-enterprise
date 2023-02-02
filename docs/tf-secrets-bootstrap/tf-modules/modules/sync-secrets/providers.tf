terraform {
  required_providers {
    kubernetes = {
      source                = "hashicorp/kubernetes"
      version               = "~> 2.14"
      configuration_aliases = [kubernetes.this, kubernetes.leaf]
    }
  }
}
