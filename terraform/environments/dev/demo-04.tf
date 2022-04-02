module "demo_04" {
  source = "../../modules/eks-cluster"

  cluster_name = "demo-04"

  cluster_tags = {
    Application = "Weave GitOps Enterprise"
    Environment = "dev"
    Team = "pesto"
  }
}