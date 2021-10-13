package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodSetMemoryLimits(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NoMemoryLimits"
    }
}

PodSetMemoryLimits(resource) {
    containers := resource.Object.spec.containers[_]
    not containers.resources.limits.memory
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    workloadsSetMemoryLimits(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NoMemoryLimits"
    }
}

workloadsSetMemoryLimits(resource) {
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.resources.limits.memory
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronjobSetMemoryLimits(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NoMemoryLimits"
    }
}

CronjobSetMemoryLimits(resource) {
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.resources.limits.memory
}