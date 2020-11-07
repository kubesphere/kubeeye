package options

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
)

type NodeProblemDetectorOptions struct{
	ServerPort int
	ServerAddress string
	NodeName string
	SystemLogMonitorConfigPaths []string
}
//func NewNodeProbelemDetectorOptions() *NodeProblemDetectorOptions{
//	npdo :=
//}

func (npdo *NodeProblemDetectorOptions) AddFlags(fs *pflag.FlagSet){
	fs.IntVar(&npdo.ServerPort, "port",
		20256, "The port to bind the node problem detector server. Use 0 to disable.")
	fs.StringVar(&npdo.ServerAddress, "address",
		"127.0.0.1", "The address to bind the node problem detector server.")

}

func (npdo *NodeProblemDetectorOptions) SetNodeName(){
	npdo.NodeName = os.Getenv("NODE_NAME")
	if npdo.NodeName != "" {
		return
	}
	nodeName, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("Failed to get host name: %v", err))
	}

	npdo.NodeName = nodeName
}

