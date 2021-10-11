package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodSetallowPrivilegeEscalation(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%' allowPrivilegeEscalation should be set false.", [resourcename])
    }
}

PodSetallowPrivilegeEscalation(resource) {
    resource.Object.spec.securityContext.allowPrivilegeEscalation == true
} else {
    resource.Object.spec.containers[_].securityContext.allowPrivilegeEscalation == true
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    DownloadsSetallowPrivilegeEscalation(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v allowPrivilegeEscalation should be set false.", [resourcename])
    }
}

DownloadsSetallowPrivilegeEscalation(resource) {
    resource.Object.spec.template.spec.containers[_].securityContext.allowPrivilegeEscalation == true
} else {
    resource.Object.spec.template.spec.securityContext.allowPrivilegeEscalation == true
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronjobSetallowPrivilegeEscalation(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v allowPrivilegeEscalation should be set false.", [resourcename])
    }
}

CronjobSetallowPrivilegeEscalation(resource) {
    resource.Object.spec.jobTemplate.spec.template.spec.containers[_].securityContext.allowPrivilegeEscalation == true
} else {
    resource.Object.spec.jobTemplate.spec.template.spec.securityContext.allowPrivilegeEscalation == true
}