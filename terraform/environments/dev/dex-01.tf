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