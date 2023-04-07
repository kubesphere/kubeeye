package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"
    level := "warning"

    PodSetlivenessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
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
    level := "warning"

    workloadsSetlivenessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
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
    level := "warning"

    CronjobSetlivenessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "NoLivenessProbe"
    }
}

CronjobSetlivenessProbe(resource) {
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.livenessProbe
}