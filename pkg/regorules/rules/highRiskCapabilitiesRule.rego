package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"
    level := "danger"

    PodSetHighRiskCapabilities(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "HighRiskCapabilities"
    }
}

PodSetHighRiskCapabilities(resource) {
    HighRiskCapabilities := ["NET_ADMIN", "SYS_ADMIN", "ALL"]
    containers := resource.Object.spec.containers[_]
    HighRiskCapabilities[_] == containers.securityContext.capabilities.add[_]
} else {
    HighRiskCapabilities := ["NET_ADMIN", "SYS_ADMIN", "ALL"]
    HighRiskCapabilities[_] == resource.Object.spec.securityContext.capabilities.add[_]
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]
    level := "danger"

    WorkloadsSetHighRiskCapabilities(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "HighRiskCapabilities"
    }
}

WorkloadsSetHighRiskCapabilities(resource) {
    HighRiskCapabilities := ["NET_ADMIN", "SYS_ADMIN", "ALL"]
    containers := resource.Object.spec.template.spec.containers[_]
    HighRiskCapabilities[_] == containers.securityContext.capabilities.add[_]
} else {
    HighRiskCapabilities := ["NET_ADMIN", "SYS_ADMIN", "ALL"]
    HighRiskCapabilities[_] == resource.Object.spec.template.spec.securityContext.capabilities.add[_]
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"
    level := "danger"

    CronjobSetHighRiskCapabilities(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "HighRiskCapabilities"
    }
}

CronjobSetHighRiskCapabilities(resource) {
    HighRiskCapabilities := ["NET_ADMIN", "SYS_ADMIN", "ALL"]
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    HighRiskCapabilities[_] == containers.securityContext.capabilities.add[_]
} else {
    HighRiskCapabilities := ["NET_ADMIN", "SYS_ADMIN", "ALL"]
    HighRiskCapabilities[_] == resource.Object.spec.jobTemplate.spec.template.spec.securityContext.capabilities.add[_]
}