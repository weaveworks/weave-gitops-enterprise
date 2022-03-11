module "demo_02" {
  source = "../../modules/gke-cluster"

  cluster_name = "demo-02"
  region = "europe-north1"
  location = "europe-north1-a"
  machine_type = "n1-standard-2"

  cluster_labels = {
    app = "wge"
    env = "dev"
    team = "pesto"
  }
}

output "demo_02_endpoint" {
  value = module.demo_02.endpoint
}
 
output "demo_02_client_certificate" {
  value = module.demo_02.client_certificate
  sensitive = true
}

output "demo_02_client_key" {
  value = module.demo_02.client_key
  sensitive = true
}

output "demo_02_cluster_ca_certificate" {
  value = module.demo_02.cluster_ca_certificate
  sensitive = true
}

output "demo_02_kubeconfig" {
  value = module.demo_02.kubeconfig
  sensitive = true
}