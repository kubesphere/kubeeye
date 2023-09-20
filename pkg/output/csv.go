package output

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/xuri/excelize/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"path"
	"strings"
)

func CSVOutput(clients *kube.KubernetesClient, outPath *string, taskName string, namespace string) error {
	filename := "kubeEyeAuditResult.xlsx"
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	list, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults().List(context.TODO(), metav1.ListOptions{
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

				if result.ResultItems != nil {
					for _, resultItem := range result.ResultItems {
						_ = f.SetSheetRow("opa", fmt.Sprintf("A%d", row), &[]string{result.NameSpace, result.ResourceType, result.Name, resultItem.Level, resultItem.Message, resultItem.Reason})
						row++
					}
				} else {
					_ = f.SetSheetRow("opa", fmt.Sprintf("A%d", row), &[]string{result.NameSpace, result.ResourceType, result.Name})
					row++
				}

			}
		}

		//if item.Spec.PrometheusResult != nil {
		//	row := 1
		//
		//	for _, prometheus := range item.Spec.PrometheusResult {
		//		var header []string
		//		for _, val := range prometheus {
		//			if len(header) == 0 {
		//				for k := range val {
		//					header = append(header, k)
		//				}
		//				_ = f.SetSheetRow("prometheus", fmt.Sprintf("A%d", row), &header)
		//				row++
		//			}
		//			var value []string
		//			for i := range header {
		//				value = append(value, val[header[i]])
		//			}
		//			_ = f.SetSheetRow("prometheus", fmt.Sprintf("A%d", row), &value)
		//			row++
		//		}
		//	}
		//
		//}

		if item.Spec.NodeInfoResult != nil {
			row := 1
			err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{"nodeName", "type", "name", "value", "assert"})
			row++
			if err != nil {
				fmt.Println(err)
			}
			for key, val := range item.Spec.NodeInfoResult {
				if val.SysctlResult != nil {
					for _, resultItem := range val.SysctlResult {
						err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{key, "sysctl", resultItem.Name, *resultItem.Value, utils.BoolToString(resultItem.Assert)})
						row++
						if err != nil {
							fmt.Println(err)
						}
					}
				}
				if val.SystemdResult != nil {
					for _, resultItem := range val.SystemdResult {
						err := f.SetSheetRow("nodeInfo", fmt.Sprintf("A%d", row), &[]string{key, "systemd", resultItem.Name, *resultItem.Value, utils.BoolToString(resultItem.Assert)})
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
