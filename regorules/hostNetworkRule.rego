package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodSethostNetwork(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "HostNetworkAllowed"
    }
}

PodSethostNetwork(resource) {
    resource.Object.spec.hostNetwork == true
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    workloadsSethostNetwork(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "HostNetworkAllowed"
    }
}

workloadsSethostNetwork(resource) {
    resource.Object.spec.template.spec.hostNetwork == true
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronJobSethostNetwork(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": "HostNetworkAllowed"
    }
}

CronJobSethostNetwork(resource) {
    resource.Object.spec.jobTemplate.spec.template.spec.hostNetwork == true
}