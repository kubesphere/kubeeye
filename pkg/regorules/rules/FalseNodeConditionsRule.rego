package kubeeye_nodes_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "Node"

    resource.Object.status.conditions[i].status == "False"
    Reason := resource.Object.status.conditions[i].reason

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Type": sprintf("%v", [type]),
        "Message": sprintf("%v", [Reason]),
    }
}