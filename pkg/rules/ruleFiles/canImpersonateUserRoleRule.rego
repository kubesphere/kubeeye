package kubeeye_RBAC_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Role"
    level := "warning"

    isNotDefaultRBAC(resource)
    canImpersonateUserResource(resource)
    haveImpersonateUserResourceVerb(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "CanImpersonateUser"
    }
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "ClusterRole"
    level := "warning"

    isNotDefaultRBAC(resource)
    canImpersonateUserResource(resource)
    haveImpersonateUserResourceVerb(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "CanImpersonateUser"
    }
}

isNotDefaultRBAC(resource) {
    not resource.Object.metadata.labels["kubernetes.io/bootstrapping"] == "rbac-defaults"
}

canImpersonateUserResource(resource){
    resource.Object.rules[_].resources[_] == "users"
}

canImpersonateUserResource(resource){
    resource.Object.rules[_].resources[_] == "groups"
}

canImpersonateUserResource(resource){
    resource.Object.rules[_].resources[_] == "serviceaccounts"
}

canImpersonateUserResource(resource){
    resource.Object.rules[_].resources[_] == "*"
}

haveImpersonateUserResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "impersonate"
}

haveImpersonateUserResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "*"
}