package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"
    level := "warning"

    PodSetreadOnlyRootFilesystem(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "NotReadOnlyRootFilesystem"
    }
}

PodSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem == true
    containers := resource.Object.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem == true
}
PodSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem
    containers := resource.Object.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem == true
}
PodSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem == true
    containers := resource.Object.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem
}
PodSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem
    containers := resource.Object.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]
    level := "warning"

    workloadsSetreadOnlyRootFilesystem(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "NotReadOnlyRootFilesystem"
    }
}

workloadsSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem == true
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem == true
}
workloadsSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem == true
}
workloadsSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem == true
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem
}
workloadsSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"
    level := "warning"

    CronJobSetreadOnlyRootFilesystem(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "NotReadOnlyRootFilesystem"
    }
}

CronJobSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem == true
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem == true
}
CronJobSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem == true
}
CronJobSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem == true
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem
}
CronJobSetreadOnlyRootFilesystem(resource) {
    not resource.Object.spec.securityContext.readOnlyRootFilesystem
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.securityContext.readOnlyRootFilesystem
}