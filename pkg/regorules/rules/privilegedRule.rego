package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodSetPrivilegedRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "PrivilegedAllowed"
    }
}

PodSetPrivilegedRule(resource) {
    resource.Object.spec.containers[_].securityContext.privileged == true
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    workloadsSetPrivilegedRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "PrivilegedAllowed"
    }
}

workloadsSetPrivilegedRule(resource) {
    resource.Object.spec.template.spec.containers[_].securityContext.privileged == true
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronJobSetPrivilegedRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "PrivilegedAllowed"
    }
}

CronJobSetPrivilegedRule(resource) {
    resource.Object.spec.jobTemplate.spec.template.spec.containers[_].securityContext.privileged == true
}