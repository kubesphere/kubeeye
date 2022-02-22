package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/pkg/errors"
)

func defaultOutput(receiver <-chan kube.ValidateResults) {
	w := tabwriter.NewWriter(os.Stdout, 10, 4, 3, ' ', 0)
	fmt.Fprintln(w, "\nKIND\tNAMESPACE\tNAME\tMESSAGE")
	for r := range receiver {
		for _, result := range r.ValidateResults {
			if len(result.Message) != 0 {
				s := fmt.Sprintf("%s\t%s\t%s\t%-8v", result.Type, result.Namespace, result.Name, result.Message)
				fmt.Fprintln(w, s)
			}
		}
	}
	w.Flush()
}

func JSONOutput(receiver <-chan kube.ValidateResults) {
	var output []kube.ResultReceiver

	for r := range receiver {
		for _, result := range r.ValidateResults {
			if len(result.Message) != 0 {
				output = append(output, result)
			}
		}
	}
	// output json
	jsonOutput, _ := json.MarshalIndent(output, "", "    ")
	fmt.Println(string(jsonOutput))
}

func CSVOutput(receiver <-chan kube.ValidateResults) {
	var output []kube.ResultReceiver
	for r := range receiver {
		for _, result := range r.ValidateResults {
			if len(result.Message) != 0 {
				output = append(output, result)
			}
		}
	}
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
	header := []string{"name", "namespace", "kind", "message", "reason"}
	data := [][]string{
		header,
	}
	for _, receiver := range output {
		var resourcename string
		for _, msg := range receiver.Message {
			if resourcename == "" {
				contexts := []string{
					receiver.Name,
					receiver.Namespace,
					receiver.Type,
					msg,
					receiver.Reason,
				}
				data = append(data, contexts)
				resourcename = receiver.Name
			} else {
				contexts := []string{
					"",
					"",
					"",
					msg,
					receiver.Reason,
				}
				data = append(data, contexts)
			}

		}
	}
	// WriteAll writes multiple CSV records to w using Write and then calls Flush,
	if err := w.WriteAll(data); err != nil {
		fmt.Println("The result is exported to kubeeyeauditResult.CSV, please check it for audit result.")
	}
}
