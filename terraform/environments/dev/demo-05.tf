# Uncomment this to create an AKS cluster
# module "demo_05" {
#   source = "../../modules/aks-cluster"

#   cluster_name = "demo-05"

#   cluster_tags = {
#     app = "wge"
#     env = "dev"
#     team = "pesto"
#   }
# }

# output "demo_05_kubeconfig" {
#   value = module.demo_05.kubeconfig
#   sensitive = true
# }