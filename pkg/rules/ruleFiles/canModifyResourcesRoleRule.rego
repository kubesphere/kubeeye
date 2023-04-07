package kubeeye_RBAC_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Role"
    level := "warning"

    isNotDefaultRBAC(resource)
    canModifyResources(resource)
    haveModifyResourcesVerb(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "CanDeleteResources"
    }
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "ClusterRole"
    level := "warning"

    isNotDefaultRBAC(resource)
    canModifyResources(resource)
    haveModifyResourcesVerb(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "CanDeleteResources"
    }
}

isNotDefaultRBAC(resource) {
    not resource.Object.metadata.labels["kubernetes.io/bootstrapping"] == "rbac-defaults"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "secrets"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "configmaps"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "persistentvolumeclaims"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "horizontalpodautoscalers"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "events"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "roles"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "clusterroles"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "users"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "groups"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "serviceaccounts"
}


canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "services"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "ingresses"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "endpoints"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "networkpolicies"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "certificates"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "certificaterequests"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "issuers"
}

canModifyResources(resource){
    resource.Object.rules[_].resources[_] == "*"
}

haveModifyResourcesVerb(resource){
    resource.Object.rules[_].verbs[_] == "create"
}

haveModifyResourcesVerb(resource){
    resource.Object.rules[_].verbs[_] == "update"
}

haveModifyResourcesVerb(resource){
    resource.Object.rules[_].verbs[_] == "patch"
}

haveModifyResourcesVerb(resource){
    resource.Object.rules[_].verbs[_] == "delete"
}

haveModifyResourcesVerb(resource){
    resource.Object.rules[_].verbs[_] == "deletecollection"
}

haveModifyResourcesVerb(resource){
    resource.Object.rules[_].verbs[_] == "*"
}