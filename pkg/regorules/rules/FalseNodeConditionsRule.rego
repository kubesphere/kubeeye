package kubeeye_nodes_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "Node"
    level := "waring"

    resource.Object.status.conditions[i].status == "False"

    contains(resource.Object.status.conditions[i].message, "has")
    not contains(resource.Object.status.conditions[i].message, "has no")
    Message := replace(resource.Object.status.conditions[i].message,"has", "has no")
    contains(resource.Object.status.conditions[i].reason, "Has")
    not contains(resource.Object.status.conditions[i].reason, "HasNo")
    Reason := replace(resource.Object.status.conditions[i].reason,"Has", "HasNo")

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": sprintf("%v", [Reason]),
        "Reason": sprintf("%v", [Message]),
    }
}

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "Node"
    level := "waring"

    resource.Object.status.conditions[i].status == "False"

    contains(resource.Object.status.conditions[i].message, "has no")
    Message := replace(resource.Object.status.conditions[i].message,"has no", "has")
    contains(resource.Object.status.conditions[i].reason, "HasNo")
    Reason := replace(resource.Object.status.conditions[i].reason,"HasNo", "Has")

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": sprintf("%v", [Reason]),
        "Reason": sprintf("%v", [Message]),
    }
}