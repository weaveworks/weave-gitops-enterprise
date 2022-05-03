variable "cluster_name" {
    description = "The name of the cluster"
}

variable "cluster_tags" {
  default = {}
  description = "The set of tags to apply to the cluster"
}

variable "node_count" {
  default     = 1
  description = "The number of nodes in the node pool"
}

variable "vm_size" {
  default     = "Standard_D2_v2"
  description = "The VM size to use for the nodes"
}

variable "kubernetes_version" {
  default = "1.23.3"
  description = "The Kubernetes version of the cluster"
}