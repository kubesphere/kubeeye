package audit

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/aquasecurity/kube-bench/check"
	"github.com/go-logr/logr"
)

type KubeBenchResponse struct {
	Controls []check.Controls
}

func KubeBenchAudit(log logr.Logger) (controls KubeBenchResponse) {
	// exec KubeBench audit, put the result into cmd
	cmd := exec.Command("kube-bench", "--json")
	// get the response from cmd.Output(), it contains the result of KubeBench audit
	response, err := cmd.Output()
	if err != nil {
		log.Error(err, "failed to execute KubeBench")
	}

	// decode the result of KubeBench audit, put the result into allControls
	decoder := json.NewDecoder(strings.NewReader(string(response)))

	err = decoder.Decode(&controls)
	if err == io.EOF {
		log.Error(err, "the result of KubeBench are empty")
	}
	if err != nil {
		log.Error(err, "failed to decode the result of KubeBench ")
	}

	return controls
}
