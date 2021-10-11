package kubeeye_RBAC_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Role"

    isNotDefaultRBAC(resource)
    canModifyPodResource(resource)
    haveModifyPodResourceVerb(resource)


    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v can modify workloads.", [resourcename])
    }
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "ClusterRole"

    isNotDefaultRBAC(resource)
    canModifyPodResource(resource)
    haveModifyPodResourceVerb(resource)


    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v can modify workloads.", [resourcename])
    }
}

isNotDefaultRBAC(resource) {
    not resource.Object.metadata.labels["kubernetes.io/bootstrapping"] == "rbac-defaults"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "pods"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "deployments"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "daemonsets"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "replicasets"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "statefulsets"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "jobs"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "cronjobs"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "replicationcontrollers"
}

canModifyPodResource(resource){
    resource.Object.rules[_].resources[_] == "*"
}

haveModifyPodResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "create"
}

haveModifyPodResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "update"
}

haveModifyPodResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "patch"
}

haveModifyPodResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "delete"
}

haveModifyPodResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "deletecollection"
}

haveModifyPodResourceVerb(resource){
    resource.Object.rules[_].verbs[_] == "*"
}