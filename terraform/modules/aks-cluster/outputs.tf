output "client_certificate" {
    description = "Public certificate used by clients to authenticate to the cluster endpoint."
    value = azurerm_kubernetes_cluster.cluster.kube_config.0.client_certificate
}

output "kubeconfig" {
    description = "A kubeconfig file configured to access the AKS cluster."
    value = azurerm_kubernetes_cluster.cluster.kube_config_raw
}