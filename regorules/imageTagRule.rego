package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodImageTagRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v image tag not specified, do not use 'latest'.", [resourcename])
    }
}

PodImageTagRule(resource) {
    regex.match("^.+:latest$", input.Object.spec.containers[_].image)
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    workloadsImageTagRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v image tag not specified, do not use 'latest'.", [resourcename])
    }
}

workloadsImageTagRule(resource) {
    regex.match("^.+:latest$", resource.Object.spec.template.spec.containers[_].image)
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronJobImageTagRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v image tag not specified, do not use 'latest'.", [resourcename])
    }
}

CronJobImageTagRule(resource) {
    regex.match("^.+:latest$", resource.Object.spec.jobTemplate.spec.template.spec.containers[_].image)
}