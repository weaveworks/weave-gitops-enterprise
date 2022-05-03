resource "azurerm_resource_group" "resource_group" {
  name     = "team-pesto"
  location = "East US"
}

resource "azurerm_kubernetes_cluster" "cluster" {
  name                = var.cluster_name
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  dns_prefix          = var.cluster_name
  kubernetes_version = var.kubernetes_version

  default_node_pool {
    name       = "default"
    node_count = var.node_count
    vm_size    = var.vm_size
  }

  network_profile {
    network_plugin    = "azure"
  }

  identity {
    type = "SystemAssigned"
  }

  tags = var.cluster_tags
}