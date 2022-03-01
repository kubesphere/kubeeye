package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"
    level := "warning"

    PodimagePullPolicyRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "ImagePullPolicyNotAlways"
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
    level := "warning"

    workloadsimagePullPolicyRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "ImagePullPolicyNotAlways"
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
    level := "warning"

    CronJobimagePullPolicyRule(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "ImagePullPolicyNotAlways"
    }
}

CronJobimagePullPolicyRule(resource) {
    imagePullPolicy := resource.Object.spec.jobTemplate.spec.template.spec.containers[_].imagePullPolicy
    imagePullPolicy != "Always"
}