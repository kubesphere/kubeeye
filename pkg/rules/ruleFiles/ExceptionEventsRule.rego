package kubeeye_events_rego

deny[msg] {
    resource := input
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    resourcenamespace := resource.Object.metadata.namespace
    type == "Event"
    level := "warning"
    Message := resource.Object.reason
    Reason := resource.Object.message

    resource.Object.type != "Normal"

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Namespace": sprintf("%v", [resourcenamespace]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": sprintf("%v", [Message]),
        "Reason": sprintf("%v", [Reason]),
    }
}