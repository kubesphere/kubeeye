package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodSetreadOnlyRootFilesystem(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v root file system should be set read only.", [resourcename])
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

    workloadsSetreadOnlyRootFilesystem(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v root file system should be set read only.", [resourcename])
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

    CronJobSetreadOnlyRootFilesystem(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v root file system should be set read only.", [resourcename])
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