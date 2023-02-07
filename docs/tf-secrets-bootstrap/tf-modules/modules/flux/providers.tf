terraform {
  required_providers {
    kubectl = {
      source                = "gavinbunney/kubectl"
      version               = "~> 1.14"
      configuration_aliases = [kubectl.this, kubectl.leaf]
    }
    flux = {
      source                = "fluxcd/flux"
      version               = ">= 0.20.0"
      configuration_aliases = [flux]
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    github = {
      source                = "integrations/github"
      version               = "~> 5.0"
      configuration_aliases = [github]
    }
  }
}
