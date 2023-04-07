package kubeeye_workloads_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Pod"
    level := "warning"

    PodSetInsecureCapabilities(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "InsecureCapabilities"
    }
}

PodSetInsecureCapabilities(resource) {
    insecureCapabilities := ["CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "NET_BIND_SERVICE","SYS_CHROOT","KILL","AUDIT_WRITE"]
    containers := resource.Object.spec.containers[_]
    insecureCapabilities[_] == containers.securityContext.capabilities.add[_]
} else {
    insecureCapabilities := ["CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "NET_BIND_SERVICE","SYS_CHROOT","KILL","AUDIT_WRITE"]
    insecureCapabilities[_] == resource.Object.spec.securityContext.capabilities.add[_]
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
    workloadsType[type]
    level := "warning"

    WorkloadsSetInsecureCapabilities(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "InsecureCapabilities"
    }
}

WorkloadsSetInsecureCapabilities(resource) {
    insecureCapabilities := ["CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "NET_BIND_SERVICE","SYS_CHROOT","KILL","AUDIT_WRITE"]
    containers := resource.Object.spec.template.spec.containers[_]
    insecureCapabilities[_] == containers.securityContext.capabilities.add[_]
} else {
    insecureCapabilities := ["CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "NET_BIND_SERVICE","SYS_CHROOT","KILL","AUDIT_WRITE"]
    insecureCapabilities[_] == resource.Object.spec.template.spec.securityContext.capabilities.add[_]
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "CronJob"
    level := "warning"

    CronjobSetInsecureCapabilities(resource)

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": "InsecureCapabilities"
    }
}

CronjobSetInsecureCapabilities(resource) {
    insecureCapabilities := ["CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "NET_BIND_SERVICE","SYS_CHROOT","KILL","AUDIT_WRITE"]
    containers := resource.Object.spec.jobTemplate.spec.template.spec.containers[_]
    insecureCapabilities[_] == containers.securityContext.capabilities.add[_]
} else {
    insecureCapabilities := ["CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER", "MKNOD", "NET_RAW", "SETGID", "SETUID", "SETFCAP", "NET_BIND_SERVICE","SYS_CHROOT","KILL","AUDIT_WRITE"]
    insecureCapabilities[_] == resource.Object.spec.jobTemplate.spec.template.spec.securityContext.capabilities.add[_]
}