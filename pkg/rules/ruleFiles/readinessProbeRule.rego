package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"
    level := "warning"

    PodSetreadinessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "NoReadinessProbe"
    }
}

PodSetreadinessProbe(resource) {
    containers := resource.Object.spec.containers[_]
    not containers.readinessProbe
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]
    level := "warning"

    workloadsSetreadinessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "NoReadinessProbe"
    }
}

workloadsSetreadinessProbe(resource) {
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.readinessProbe
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"
    level := "warning"

    CronjobSetreadinessProbe(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "NoReadinessProbe"
    }
}

CronjobSetreadinessProbe(resource) {
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.readinessProbe
}