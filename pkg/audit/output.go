package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	"github.com/pkg/errors"
)

func defaultOutput(receiver <-chan v1alpha1.AuditResult) {
	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	fmt.Fprintln(w, "\nKIND\tNAMESPACE\tNAME\tREASON\tLEVEL\tMESSAGE")
	for r := range receiver {
		for _, results := range r.Results {
			for _, resultInfos := range results.ResultInfos {
				for _, resourceInfos := range resultInfos.ResourceInfos {
					for _, items := range resourceInfos.ResultItems {
						if len(items.Message) != 0 {
							s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%-8v", results.ResourcesType, resultInfos.Namespace,
								resourceInfos.Name, items.Reason, items.Level, items.Message)
							fmt.Fprintln(w, s)
						}
					}
				}
			}
		}
	}
	w.Flush()
}

func JSONOutput(receiver <-chan v1alpha1.AuditResult) {
	var output v1alpha1.AuditResult
	for r := range receiver {
		for _, results := range r.Results {
			output.Results = append(output.Results, results)
		}
	}

	// output json
	jsonOutput, _ := json.MarshalIndent(output, "", "    ")
	fmt.Println(string(jsonOutput))
}

func CSVOutput(receiver <-chan v1alpha1.AuditResult) {
	filename := "kubeEyeAuditResult.csv"
	// create csv file
	newFile, err := os.Create(filename)
	if err != nil {
		createError := errors.Wrap(err, "create file kubeEyeAuditResult.csv failed.")
		panic(createError)
	}

	defer newFile.Close()

	// write UTF-8 BOM to prevent print gibberish.
	newFile.WriteString("\xEF\xBB\xBF")

	// NewWriter returns a new Writer that writes to w.
	w := csv.NewWriter(newFile)
	header := []string{"namespace", "kind", "name", "level", "message", "reason"}
	contents := [][]string{
		header,
	}
	for r := range receiver {
		for _, results := range r.Results {
			for _, resultInfos := range results.ResultInfos {
				var resourceName string
				for _, resourceInfos := range resultInfos.ResourceInfos {
					for _, items := range resourceInfos.ResultItems {
						if resourceName == "" {
							content := []string{
								resultInfos.Namespace,
								results.ResourcesType,
								resourceInfos.Name,
								items.Level,
								items.Message,
								items.Reason,
							}
							contents = append(contents, content)
							resourceName = resourceInfos.Name
						} else {
							content := []string{
								"",
								"",
								"",
								items.Level,
								items.Message,
								items.Reason,
							}
							contents = append(contents, content)
						}
					}
				}
			}
		}
	}
	// WriteAll writes multiple CSV records to w using Write and then calls Flush,
	if err := w.WriteAll(contents); err != nil {
		fmt.Println("The result is exported to kubeEyeAuditResult.CSV, please check it for audit result.")
	}
}
