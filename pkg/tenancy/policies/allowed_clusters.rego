package weave.tenancy.allowed_clusters

controller_input := input.review.object
namespace := controller_input.metadata.namespace
secrets := input.parameters.cluster_secrets
violation[result] {
    controller_input.kind == "GitopsCluster"
    secret := controller_input.spec.secretRef.name
    not contains_array(secret, secrets)
    result = {
    "issue detected": true,
    "msg": sprintf("cluster secretRef %v is not allowed for namespace %v", [secret, namespace]),
    }
}
violation[result] {
    controller_input.kind == "Kustomization"
    secret := controller_input.spec.kubeConfig.secretRef.name
    not contains_array(secret, secrets)
    result = {
    "issue detected": true,
    "msg": sprintf("kustomization kubeconfig secretRef %v is not allowed for namespace %v", [secret, namespace]),
    }
}
contains_array(item, items) {
    items[_] = item
}
