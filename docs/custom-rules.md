## Add your own command
``` text
├── cmd
│   └── testcmd.go
├── main.go
```
testcmd.go
```go
package cmd
import (
	"fmt"
	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "test",
	Long:  `new command`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("new command")
	},
}
```
main.go
``` go
package main

import (
	"github.com/leonharetd/kubeeye/cmd"
	kc "github.com/leonharetd/kubeeye_sample/cmd"
)

func main() {
	cmd := cmd.NewKubeEyeCommand().WithCommand(kc.TestCmd).DO()
	cmd.Execute()
}
```
Use after build
```shell
>> kubeeye audit
KubeEye finds various problems on Kubernetes cluster.

Usage:
  ke [command]

Available Commands:
  audit       audit resources from the cluster
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  install     A brief description of your command
  test        test
  uninstall   A brief description of your command
```
## Add your own custom rules
### Embed custom OPA rules
``` text
├── main.go
└── regorules
    ├── rules
    │   ├── imageRegistryRule.rego
    │   └── testRule.rego
    └── testrule.go
```
testrule.go

specify embed directory
``` go
package regorules

import (
	"embed"
)

//go:embed rules
var EmbRegoRules embed.FS
```
``` go
package main

import (
	"github.com/leonharetd/kubeeye/cmd"
	"github.com/leonharetd/kubeeye_sample/regorules"
)

func main() {
	cmd := cmd.NewKubeEyeCommand().WithRegoRule(regorules.EmbRegoRules).DO()
	cmd.Execute()
}
```
If you have multiple directories
``` go 
cmd := cmd.NewKubeEyeCommand().WithRegoRule(RulesA).WithRegoRule(RulesB).DO()
```
Use after build
```shell
kubeeye audit
```
### Embed custom function rules
github.com/leonharetd/kubeeye_sample/expirerules/expirerule.go
```go
package funcrules

import (
	"fmt"
	"strconv"
	kube "github.com/leonharetd/kubeeye/pkg/kube"
)

type ExpireTestRule struct{}

func (cer ExpireTestRule) Exec() kube.ValidateResults {
	output := kube.ValidateResults{ValidateResults: make([]kube.ResultReceiver, 0)}
	var certExpiresOutput kube.ResultReceiver
	for i := range []int{1, 2, 3, 4} {
		certExpiresOutput.Name = fmt.Sprint("test", strconv.Itoa(i))
		certExpiresOutput.Type = "testExpire"
		certExpiresOutput.Message = []string{strconv.Itoa(i), "expire"}
		output.ValidateResults = append(output.ValidateResults, certExpiresOutput)
	}
	return output
}
```
main.go
``` go
package main

import (
	"github.com/leonharetd/kubeeye/cmd"
	"github.com/leonharetd/kubeeye_sample/funcrules"
)

func main() {
	cmd := cmd.NewKubeEyeCommand().WithFuncRule(funcrules.FuncTestRule{}).DO()
	cmd.Execute()
}
```
Use after build
```shell
kubeeye audit
```