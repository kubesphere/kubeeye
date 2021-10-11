package kube

import (
    "fmt"
    "io/ioutil"
    "os"
    "strings"
)


func GetRegoRules(additionalRegoRulePath string)  {
    var regoRulesList RegoRulesList
    if additionalRegoRulePath != "" {
        GetRegoRulesfiles(additionalRegoRulePath, &regoRulesList)
    }
     GetRegoRulesfiles("./regorules/", &regoRulesList)

    RegoRulesListChan <- regoRulesList
}

func GetRegoRulesfiles(path string, regoRulesList *RegoRulesList) {
    var regoRules []string
    if strings.HasSuffix(path,"/") == false{
        path += "/"
    }
    files, err := ioutil.ReadDir(path)
    if err != nil {
        fmt.Println("Failed to read the dir of rego rule files.")
        os.Exit(1)
    }

    for _, file := range files {
        if strings.HasSuffix(file.Name(),".rego") == false{
            continue
        }

        getregoRule, _ := ioutil.ReadFile(path + file.Name())
        regoRule := string(getregoRule)
        regoRules = append(regoRules,regoRule)
    }
    regoRulesList.RegoRules = append(regoRulesList.RegoRules,regoRules...)
}
