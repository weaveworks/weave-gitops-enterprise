module "dex_01" {
  source = "../../modules/gke-cluster"

  cluster_name = "dex-01"
  region = "europe-north1"
  location = "europe-north1-a"
  machine_type = "n1-standard-2"

  cluster_labels = {
    app = "dex"
    env = "dev"
    team = "pesto"
  }
}

output "dex_01_endpoint" {
  value = module.dex_01.endpoint
}
 
output "dex_01_client_certificate" {
  value = module.dex_01.client_certificate
  sensitive = true
}

output "dex_01_client_key" {
  value = module.dex_01.client_key
  sensitive = true
}

output "dex_01_cluster_ca_certificate" {
  value = module.dex_01.cluster_ca_certificate
  sensitive = true
}