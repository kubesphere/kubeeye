package create

import (
	"fmt"
	kubeeyetemplate "github.com/kubesphere/kubeeye-v1alpha2/pkg/template"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
	"path"
	"strings"
	"text/template"
)

type ConfigOptions struct {
	Template string
	Path     string
}

func NewConfigCmd() *cobra.Command {
	c := &ConfigOptions{}
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Generate Inspect Rule Config",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.GenerateInspectConfig()
			if err != nil {
				klog.Errorf("failed to generate config ,err:%s", err)
				return err
			}
			klog.Info("generate config success")
			return nil
		},
	}

	c.addFlags(configCmd)
	return configCmd
}

func (c *ConfigOptions) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.Template, "template", "", "Generate Inspect rule  (InspectRule or InspectPlan). default Generate All")
	cmd.Flags().StringVarP(&c.Path, "path", "o", "", "Generate Config output path")
}

func (c *ConfigOptions) GenerateInspectConfig() error {
	if c.Template == "" {
		err := c.GenerateInspectRule()

		if err != nil {
			return err
		}
		err = c.GenerateInspectPlan()
		if err != nil {
			return err
		}
	}
	switch strings.ToUpper(c.Template) {
	case "INSPECTRULE":
		err := c.GenerateInspectRule()
		if err != nil {
			return err
		}
	case "INSPECTPLAN":
		err := c.GenerateInspectPlan()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigOptions) GenerateInspectRule() error {
	ruleTemplate, err := kubeeyetemplate.GetInspectRuleTemplate()
	if err != nil {
		return err
	}
	err = c.RenderConfigFile(ruleTemplate, nil)
	if err != nil {
		return err
	}
	return nil
}
func (c *ConfigOptions) GenerateInspectPlan() error {
	ruleTemplate, err := kubeeyetemplate.GetInspectPlanTemplate()
	if err != nil {
		return err
	}
	err = c.RenderConfigFile(ruleTemplate, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigOptions) RenderConfigFile(temp *template.Template, data map[string]interface{}) error {
	name := c.GetFileName(temp)
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	err = temp.Execute(file, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigOptions) GetFileName(temp *template.Template) string {
	if c.Path == "" {
		return fmt.Sprintf("%s.yaml", temp.Name())
	}
	return path.Join(c.Path, fmt.Sprintf("%s.yaml", temp.Name()))

}
