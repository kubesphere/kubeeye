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

func defaultOutput(receiver <-chan v1alpha1.AuditResult) error {
	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	_, err := fmt.Fprintln(w, "\nKIND\tNAMESPACE\tNAME\tREASON\tLEVEL\tMESSAGE")
	if err != nil {
		return err
	}
	for r := range receiver {
		for _, results := range r.Results {
			for _, resultInfos := range results.ResultInfos {
				for _, resourceInfos := range resultInfos.ResourceInfos {
					for _, items := range resourceInfos.ResultItems {
						if len(items.Message) != 0 {
							s := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%-8v", results.ResourcesType, resultInfos.Namespace,
								resourceInfos.Name, items.Reason, items.Level, items.Message)
							_, err := fmt.Fprintln(w, s)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func JSONOutput(receiver <-chan v1alpha1.AuditResult) error {
	var output v1alpha1.AuditResult
	for r := range receiver {
		for _, results := range r.Results {
			output.Results = append(output.Results, results)
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

func CSVOutput(receiver <-chan v1alpha1.AuditResult) error {
	filename := "kubeEyeAuditResult.csv"
	// create csv file
	newFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "create file kubeEyeAuditResult.csv failed.")
	}

	defer newFile.Close()

	// write UTF-8 BOM to prevent print gibberish.
	if _, err = newFile.WriteString("\xEF\xBB\xBF"); err != nil {
		return err
	}

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
		return err
	}
	fmt.Printf("\033[1;36;49mThe result is exported to kubeEyeAuditResult.CSV, please check it for audit result.\033[0m\n")
	return nil
}
