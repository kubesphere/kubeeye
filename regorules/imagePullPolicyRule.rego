package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"

    PodimagePullPolicyRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v imagePullPolicy should be set 'Always'.", [resourcename])
    }
}

PodimagePullPolicyRule(resource) {
    imagePullPolicy := resource.Object.spec.containers[_].imagePullPolicy
    imagePullPolicy != "Always"
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]

    workloadsimagePullPolicyRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v imagePullPolicy should be set 'Always'.", [resourcename])
    }
}

workloadsimagePullPolicyRule(resource) {
    imagePullPolicy := resource.Object.spec.template.spec.containers[_].imagePullPolicy
    imagePullPolicy != "Always"
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"

    CronJobimagePullPolicyRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v imagePullPolicy should be set 'Always'.", [resourcename])
    }
}

CronJobimagePullPolicyRule(resource) {
    imagePullPolicy := resource.Object.spec.jobTemplate.spec.template.spec.containers[_].imagePullPolicy
    imagePullPolicy != "Always"
}