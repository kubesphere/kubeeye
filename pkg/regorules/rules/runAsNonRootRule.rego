package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodSetrunAsNonRoot(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NotRunAsNonRoot"
    }
}

PodSetrunAsNonRoot(resource) {
    not resource.Object.spec.securityContext.runAsNonRoot
    containers := resource.Object.spec.containers[_]
    not containers.securityContext.runAsNonRoot
}
PodSetrunAsNonRoot(resource) {
    resource.Object.spec.securityContext.runAsNonRoot == false
    containers := resource.Object.spec.containers[_]
    not containers.securityContext.runAsNonRoot
}
PodSetrunAsNonRoot(resource) {
    not resource.Object.spec.securityContext.runAsNonRoot
    containers := resource.Object.spec.containers[_]
    containers.securityContext.runAsNonRoot == false
}
PodSetrunAsNonRoot(resource) {
    resource.Object.spec.securityContext.runAsNonRoot = false
    containers := resource.Object.spec.containers[_]
    containers.securityContext.runAsNonRoot == false
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    workloadsSetrunAsNonRoot(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NotRunAsNonRoot"
    }
}

workloadsSetrunAsNonRoot(resource) {
    not resource.Object.spec.securityContext.runAsNonRoot
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.securityContext.runAsNonRoot
}
workloadsSetrunAsNonRoot(resource) {
    resource.Object.spec.securityContext.runAsNonRoot == false
    containers := resource.Object.spec.template.spec.containers[_]
    not containers.securityContext.runAsNonRoot
}
workloadsSetrunAsNonRoot(resource) {
    not resource.Object.spec.securityContext.runAsNonRoot
    containers := resource.Object.spec.template.spec.containers[_]
    containers.securityContext.runAsNonRoot == false
}
workloadsSetrunAsNonRoot(resource) {
    resource.Object.spec.securityContext.runAsNonRoot = false
    containers := resource.Object.spec.template.spec.containers[_]
    containers.securityContext.runAsNonRoot == false
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronJobSetrunAsNonRoot(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "NotRunAsNonRoot"
    }
}

CronJobSetrunAsNonRoot(resource) {
    not resource.Object.spec.securityContext.runAsNonRoot
    containers := resource.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.securityContext.runAsNonRoot
}
CronJobSetrunAsNonRoot(resource) {
    resource.Object.spec.securityContext.runAsNonRoot == false
    containers := resource.spec.jobTemplate.spec.template.spec.containers[_]
    not containers.securityContext.runAsNonRoot
}
CronJobSetrunAsNonRoot(resource) {
    not resource.Object.spec.securityContext.runAsNonRoot
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    containers.securityContext.runAsNonRoot == false
}
CronJobSetrunAsNonRoot(resource) {
    resource.Object.spec.securityContext.runAsNonRoot = false
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    containers.securityContext.runAsNonRoot == false
}