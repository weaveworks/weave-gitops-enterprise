package weave.tenancy.allowed_repositories

controller_input := input.review.object
namespace := controller_input.metadata.namespace
violation[result] {
    controller_input.kind == "GitRepository"
    urls := input.parameters.git_urls
    url := controller_input.spec.url
    not contains_array(url, urls)
    result = {
    "issue detected": true,
    "msg": sprintf("Git repository url %v is not allowed for namespace %v", [url, namespace]),
    }
}
violation[result] {
    controller_input.kind == "Bucket"
    urls := input.parameters.bucket_endpoints
    url := controller_input.spec.endpoint
    not contains_array(url, urls)
    result = {
    "issue detected": true,
    "msg": sprintf("Bucket endpoint %v is not allowed for namespace %v", [url, namespace]),
    }
}
violation[result] {
    controller_input.kind == "HelmRepository"
    urls := input.parameters.helm_urls
    url := controller_input.spec.url
    not contains_array(url, urls)
    result = {
    "issue detected": true,
    "msg": sprintf("Helm repository url %v is not allowed for namespace %v", [url, namespace]),
    }
}
contains_array(item, items) {
    items[_] = item
}
