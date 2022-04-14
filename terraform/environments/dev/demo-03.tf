module "demo_03" {
  source = "../../modules/gke-cluster"

  cluster_name = "demo-03"
  region = "europe-north1"
  location = "europe-north1-a"
  machine_type = "n2-standard-2"

  cluster_labels = {
    app = "wge"
    env = "dev"
    team = "pesto"
  }
}

output "demo_03_kubeconfig" {
  value = module.demo_03.kubeconfig
  sensitive = true
}