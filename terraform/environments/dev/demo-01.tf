module "demo_01" {
  source = "../../modules/gke-cluster"

  cluster_name = "demo-01"
  region = "europe-north1"
  location = "europe-north1-a"
  machine_type = "n2-standard-2"

  cluster_labels = {
    app = "wge"
    env = "dev"
    team = "pesto"
  }
}

output "demo_01_endpoint" {
  value = module.demo_01.endpoint
}
 
output "demo_01_client_certificate" {
  value = module.demo_01.client_certificate
  sensitive = true
}

output "demo_01_client_key" {
  value = module.demo_01.client_key
  sensitive = true
}

output "demo_01_cluster_ca_certificate" {
  value = module.demo_01.cluster_ca_certificate
  sensitive = true
}

output "demo_01_kubeconfig" {
  value = module.demo_01.kubeconfig
  sensitive = true
}