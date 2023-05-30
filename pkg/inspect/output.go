package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/xuri/excelize/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	path "path"
	"strings"
	"text/tabwriter"

	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
)

func defaultOutput(receiver <-chan []v1alpha2.ResourceResult) error {
	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	_, err := fmt.Fprintln(w, "\nNAMESPACE\tKIND\tNAME\tLEVEL\tMESSAGE\tREASON")
	if err != nil {
		return err
	}
	for r := range receiver {
		for _, results := range r {
			for _, items := range results.ResultItems {
				s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%-8v", results.NameSpace, results.ResourceType,
					results.Name, items.Level, items.Message, items.Reason)
				_, err := fmt.Fprintln(w, s)
				if err != nil {
					return err
				}
			}
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func JSONOutput(receiver <-chan []v1alpha2.ResourceResult) error {
	var output []v1alpha2.ResourceResult
	for r := range receiver {
		for _, results := range r {
			output = append(output, results)
		}
	}

	// output json
	jsonOutput, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonOutput))
	return nil
}

func CSVOutput(clients *kube.KubernetesClient, outPath *string, taskName string, namespace string) error {
	filename := "kubeEyeAuditResult.xlsx"
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	list, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(map[string]string{constant.LabelName: taskName})),
	})
	if err != nil {
		klog.Error(err)
		return err
	}

	_, err = f.NewSheet("opa")
	if err != nil {
		klog.Error(err)
		return err
	}
	_, err = f.NewSheet("prometheus")
	if err != nil {
		return err
	}
	_, err = f.NewSheet("nodeInfo")
	if err != nil {
		klog.Error(err)
		return err
	}

	for _, item := range list.Items {

		if item.Spec.OpaResult.ResourceResults != nil {
			row := 1
			_ = f.SetSheetRow("opa", fmt.Sprintf("A%d", row), &[]string{"NameSpace", "Kind", "Name", "Level", "Message", "Reason"})
			row++
			for _, result := range item.Spec.OpaResult.ResourceResults {
				value := []string{result.NameSpace, result.ResourceType, result.Name}

				if result.ResultItems != nil {
					for _, resultItem := range result.ResultItems {
						value = append(value, resultItem.Level, resultItem.Message, resultItem.Reason)
						_ = f.SetSheetRow("opa", fmt.Sprintf("A%d", row), &value)
						row++
					}
				} else {
					_ = f.SetSheetRow("opa", fmt.Sprintf("A%d", row), &value)
					row++
				}

			}
		}

		if item.Spec.PrometheusResult != nil {
			row := 1

			for _, prometheus := range item.Spec.PrometheusResult {
				var header []string
				for _, val := range prometheus {
					if len(header) == 0 {
						for k := range val {
							header = append(header, k)
						}
						_ = f.SetSheetRow("prometheus", fmt.Sprintf("A%d", row), &header)
						row++
					}
					var value []string
					for i := range header {
						value = append(value, val[header[i]])
					}
					_ = f.SetSheetRow("prometheus", fmt.Sprintf("A%d", row), &value)
					row++
				}
			}

		}

		if item.Spec.NodeInfoResult != nil {
			row := 1
			err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{"nodeName", "type", "name", "value", "assert"})
			row++
			if err != nil {
				fmt.Println(err)
			}
			for key, val := range item.Spec.NodeInfoResult {
				if val.NodeInfo != nil {
					for k, v := range val.NodeInfo {
						err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{key, "nodeInfo", k, v})
						row++
						if err != nil {
							fmt.Println(err)
						}
					}
				}
				if val.SysctlResult != nil {
					for _, resultItem := range val.SysctlResult {
						err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{key, "sysctl", resultItem.Name, *resultItem.Value, utils.FormatBool(resultItem.Assert)})
						row++
						if err != nil {
							fmt.Println(err)
						}
					}
				}
				if val.SystemdResult != nil {
					for _, resultItem := range val.SystemdResult {
						err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{key, "systemd", resultItem.Name, *resultItem.Value, utils.FormatBool(resultItem.Assert)})
						row++
						if err != nil {
							fmt.Println(err)
						}
					}
				}
				if val.FileChangeResult != nil {
					for _, fileItem := range val.FileChangeResult {
						err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{key, "filechange", fileItem.FileName, strings.Join(fileItem.Issues, ",")})
						row++
						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
		}
	}
	f.SetActiveSheet(1)
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
	}

	if outPath != nil {
		filename = path.Join(*outPath, filename)
	}

	err = f.SaveAs(filename)
	if err != nil {
		return err
	}
	fmt.Printf("The result is exported to kubeEyeAuditResult.xlsx, please check it for inspect result.\n")
	return nil
}
