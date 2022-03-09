variable "region" {
  default     = "europe-west1"
  description = "GCP region"
}

variable "location" {
  default     = "europe-west1-b"
  description = "GCP location"
}

variable "cluster_name" {
    description = "The name of the cluster"
}

variable "node_count" {
  default     = 1
  description = "The number of the GKE nodes per region"
}

variable "machine_type" {
  default = "n1-standard-1"
  description = "GCP machine type"
}

variable "cluster_labels" {
  default = {}
  description = "The set of labels to apply to the cluster"
}