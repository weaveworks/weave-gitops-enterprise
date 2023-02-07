resource "kubernetes_namespace" "flux_system" {
  metadata {
    name = "flux-system"
  }


  lifecycle {
    ignore_changes = [
      metadata[0].labels,
    ]
  }
}

data "kubernetes_secret_v1" "secret-to-sync-remote" {
  metadata {
    name      = "flux-system"
    namespace = "flux-system"
  }
  provider = kubernetes.this
  depends_on = [
    kubernetes_namespace.flux_system
  ]
}

resource "kubernetes_secret_v1" "target-to-sync-remote" {
  metadata {
    name      = data.kubernetes_secret_v1.secret-to-sync-remote.metadata[0].name
    namespace = data.kubernetes_secret_v1.secret-to-sync-remote.metadata[0].namespace
  }

  data     = data.kubernetes_secret_v1.secret-to-sync-remote.data
  provider = kubernetes.leaf
}
