terraform {
  required_providers {
    kubectl = {
      source                = "gavinbunney/kubectl"
      version               = "~> 1.14"
      configuration_aliases = [kubectl.this, kubectl.leaf]
    }
  }
}
