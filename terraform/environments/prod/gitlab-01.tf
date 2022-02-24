module "gitlab_01" {
  source = "../../modules/gke-cluster"

  cluster_name = "gitlab-01"
  region = var.region
  location = "europe-west1-b"
  machine_type = "n2-standard-4"
}

output "gitlab_01_endpoint" {
  value = module.gitlab_01.endpoint
}
 
output "gitlab_01_client_certificate" {
  value = module.gitlab_01.client_certificate
  sensitive = true
}

output "gitlab_01_client_key" {
  value = module.gitlab_01.client_key
  sensitive = true
}

output "gitlab_01_cluster_ca_certificate" {
  value = module.gitlab_01.cluster_ca_certificate
  sensitive = true
}