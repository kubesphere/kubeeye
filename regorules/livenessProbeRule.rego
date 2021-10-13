package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodSetlivenessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NoLivenessProbe"
    }
}

PodSetlivenessProbe(resource) {
    containers := resource.Object.spec.containers[_]
    not containers.livenessProbe
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    workloadsSetlivenessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NoLivenessProbe"
    }
}

workloadsSetlivenessProbe(resource) {
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.livenessProbe
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronjobSetlivenessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NoLivenessProbe"
    }
}

CronjobSetlivenessProbe(resource) {
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.livenessProbe
}