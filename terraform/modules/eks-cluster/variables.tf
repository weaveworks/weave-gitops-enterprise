variable "cluster_name" {
    description = "The name of the cluster"
}

variable "cluster_version" {
    default     = "1.21"
    description = "The Kubernetes version"
}

variable "desired_size" {
    default     = 1
    description = "The desired number of nodes in the EKS managed group"
}

variable "cluster_tags" {
    description = "Tags to associate with the cluster"
}

variable "oidc_client_id" {
    default     = "kubernetes-oidc-login"
    description = "The client ID to use for the OIDC flow"
}

variable "oidc_issuer_url" {
    default     = "https://dex-01.wge.dev.weave.works"
    description = "The OIDC issuer URL"
}

variable "oidc_identity_provider_config_name" {
    default     = "dex-01"
    description = "The name of the OIDC configuration"
}