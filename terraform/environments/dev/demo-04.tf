module "demo_04" {
  source = "../../modules/eks-cluster"

  cluster_name = local.cluster_name
  cluster_tags = local.cluster_tags
}

output "demo_04_kubeconfig" {
  value = local.kubeconfig
  sensitive = true
}

locals {
  cluster_name = "demo-04"
  cluster_tags = {
    Application = "Weave GitOps Enterprise"
    Environment = "dev"
    Team = "pesto"
  }
  kubeconfig = templatefile("${path.module}/../../modules/eks-cluster/templates/kubeconfig.yaml.tftpl", {
    cluster_endpoint = module.demo_04.cluster_endpoint
    cluster_name = local.cluster_name
    user_name = local.cluster_name
    context = local.cluster_name
    cluster_certificate_authority_data = module.demo_04.cluster_certificate_authority_data
  })
}