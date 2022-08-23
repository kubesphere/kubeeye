package kubeeye_nodes_rego

deny[msg] {

    resource := input
    conditiontypes := ["MemoryPressure","DiskPressure","PIDPressure","NetworkUnavailable"]
    
    
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "Node"
    level := "waring"

    resource.Object.status.conditions[i].status == "True"
    contains_element(conditiontypes,resource.Object.status.conditions[i].type)

    
    Message := resource.Object.status.conditions[i].message
    Reason := resource.Object.status.conditions[i].reason

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
    conditiontypes := ["Ready"]
    
    
    type := resource.Object.kind
    resourcename := resource.Object.metadata.name
    type == "Node"
    level := "waring"

    resource.Object.status.conditions[i].status != "True"
    contains_element(conditiontypes,resource.Object.status.conditions[i].type)

    
    Message := resource.Object.status.conditions[i].message
    Reason := resource.Object.status.conditions[i].reason

    msg := {
        "Name": sprintf("%v", [resourcename]),
        "Type": sprintf("%v", [type]),
        "Level": sprintf("%v", [level]),
        "Message": sprintf("%v", [Reason]),
        "Reason": sprintf("%v", [Message]),
    }
}

contains_element(arr, elem) = true {
  arr[_] = elem
} else = false { true }