package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"
    level := "danger"

    PodSetallowPrivilegeEscalation(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "PrivilegeEscalationAllowed"
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
    level := "danger"

    DownloadsSetallowPrivilegeEscalation(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "PrivilegeEscalationAllowed"
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
    level := "danger"

    CronjobSetallowPrivilegeEscalation(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "PrivilegeEscalationAllowed"
    }
}

CronjobSetallowPrivilegeEscalation(resource) {
    resource.Object.spec.jobTemplate.spec.template.spec.containers[_].securityContext.allowPrivilegeEscalation == true
} else {
    resource.Object.spec.jobTemplate.spec.template.spec.securityContext.allowPrivilegeEscalation == true
}