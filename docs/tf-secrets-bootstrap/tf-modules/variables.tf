
variable "cluster_name" {
  type        = string
  description = "cluster name"
  default     = "<cluster-name>"
}

variable "this_cluster_name" {
  type        = string
  description = "cluster name"
  default     = "<management-cluster-name>"
}

variable "github_owner" {
  type        = string
  description = "github owner"
  default     = "<github-owner>"
}

variable "github_token" {
  type        = string
  description = "github token secret name"
  default     = "<github-token>"
}

variable "repository_name" {
  type        = string
  description = "github repository name"
  default     = "<repo-name>"
}

variable "repository_visibility" {
  type        = string
  description = "How visible is the github repo public or private"
  default     = "private"
}

variable "branch" {
  type        = string
  description = "branch name"
  default     = "<branch-name>"
}

variable "token" {
  type        = string
  description = "cluster token"
  default     = "<cluster-name>-kubeconfig"
}

variable "target_path" {
  type        = string
  description = "flux sync target path"
  default     = "<target-path>"
}
