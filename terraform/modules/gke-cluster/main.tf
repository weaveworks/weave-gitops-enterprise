resource "google_container_cluster" "cluster" {
  name     = var.cluster_name
  location = var.location
  
  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1

  master_auth {
    client_certificate_config {
      issue_client_certificate = true
    }
  }

  resource_labels = var.cluster_labels

  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
}

resource "google_container_node_pool" "cluster_nodes" {
  name       = "${google_container_cluster.cluster.name}-node-pool"
  location   = var.location
  cluster    = google_container_cluster.cluster.name
  node_count = var.node_count

  node_config {
    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]

    # labels = {
    #   env = var.project_id
    # }

    # preemptible  = true
    machine_type = var.machine_type
    tags         = ["gke-node", "${var.cluster_name}-gke"]
    metadata = {
      disable-legacy-endpoints = "true"
    }
  }
}

# VPC
resource "google_compute_network" "vpc" {
  name                    = "${var.cluster_name}-vpc"
  auto_create_subnetworks = "false"
}

# Subnet
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.cluster_name}-subnet"
  region        = var.region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.10.0.0/24"
}

data "google_client_config" "provider" {}

data "template_file" "kubeconfig" {
  template = file("${path.module}/templates/kubeconfig.yaml.tpl")

  vars = {
    cluster_name            = google_container_cluster.cluster.name
    user_name               = google_container_cluster.cluster.name
    context                 = google_container_cluster.cluster.name
    cluster_ca_certificate  = google_container_cluster.cluster.master_auth.0.cluster_ca_certificate
    // The cluster's Kubernetes API endpoint
    endpoint                = google_container_cluster.cluster.endpoint
    // The OAuth2 access token used by the client to authenticate against the Google Cloud API.
    token                   = data.google_client_config.provider.access_token
  }
}