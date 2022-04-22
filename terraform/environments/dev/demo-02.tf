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

output "demo_02_kubeconfig" {
  value = module.demo_02.kubeconfig
  sensitive = true
}