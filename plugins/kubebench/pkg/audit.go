package pkg

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func KubeBenchAudit() (err error, results KubeBenchResults) {
	// exec KubeBench audit, put the result into cmd
	cmd := exec.Command("kube-bench", "--json")
	// get the response from cmd.Output(), it contains the result of KubeBench audit
	response, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "failed to execute KubeBench"), results
	}

	// decode the result of KubeBench audit, put the result into allControls
	decoder := json.NewDecoder(strings.NewReader(string(response)))

	err = decoder.Decode(&results)
	if err == io.EOF {
		return errors.Wrap(err, "the result of KubeBench are empty"), results
	}
	if err != nil {
		return errors.Wrap(err, "failed to decode the result of KubeBench"), results
	}

	return nil, results
}
