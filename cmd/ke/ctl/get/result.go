package get

import (
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
)

type ResultConfig struct {
	Clients       *kube.KubernetesClient
	Path          string
	Type          string
	TaskName      string
	TaskNameSpace string
}

func NewResultCmd(client *kube.KubernetesClient) *cobra.Command {
	r := &ResultConfig{
		Clients: client,
	}
	resultCmd := &cobra.Command{
		Use:     "result",
		Short:   "out inspect result to file",
		Example: "ke get result inspectTask-123456789 [flags]",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := r.CheckArgs(args)
			if err != nil {
				return err
			}
			if r.Type == "json" {
				err = output.JsonOut(cmd.Context(), r.Clients, r.Path, r.TaskName)
				if err != nil {
					klog.Error(err)
					os.Exit(1)
				}
			} else {
				err = output.HtmlOut(cmd.Context(), r.Clients, r.Path, r.TaskName)
				if err != nil {
					klog.Error(err)
					os.Exit(1)
				}
			}

			fmt.Println("result output success")
			return nil
		},
	}

	r.AddFlags(resultCmd)
	return resultCmd
}

func (r *ResultConfig) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&r.Path, "path", "p", "", "Get the output result path")
	cmd.Flags().StringVarP(&r.Type, "output", "o", "", "Get the output result type")
}

func (r *ResultConfig) CheckArgs(args []string) error {
	if len(args) < 1 {
		return errors.New("Unable to get task results")
	}
	if len(args) > 1 {
		a := args[:1]
		return errors.Errorf("invalid parameter '%s'", a[0])
	}
	r.TaskName = args[0]
	return nil
}
