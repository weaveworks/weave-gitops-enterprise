package weave.tenancy.allowed_application_deploy

controller_input := input.review.object
violation[result] {
    namespaces := input.parameters.namespaces
    targetNamespace := controller_input.spec.targetNamespace
    not contains_array(targetNamespace, namespaces)
    result = {
    "issue detected": true,
    "msg": sprintf("using target namespace %v is not allowed", [targetNamespace]),
    }
}
violation[result] {
    serviceAccountName := controller_input.spec.serviceAccountName
    serviceAccountName != input.parameters.service_account_name
    result = {
    "issue detected": true,
    "msg": sprintf("using service account name %v is not allowed", [serviceAccountName]),
    }
}
contains_array(item, items) {
    items[_] = item
}
