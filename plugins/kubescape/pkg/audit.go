package pkg

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/armosec/opa-utils/reporthandling"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

func KubescapeAudit(logs logr.Logger) (err error, auditResults []reporthandling.FrameworkReport) {
	logs.Info("start kubescape audit")
	cmd := exec.Command("kubescape", "scan", "-e", "kube-system", "-f", "json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err, auditResults
	}
	if err := cmd.Start(); err != nil {
		return err, nil
	}

	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if strings.Contains(line, "controlReports") && strings.Contains(line, "services") {
			logs.Info("decode result")
			err := json.NewDecoder(strings.NewReader(line)).Decode(&auditResults)
			if err != nil {
				return errors.Wrap(err, "decode result failed"), auditResults
			}
		}
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "get results failed"), auditResults
		}

	}

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "the command kube-hunter exec failed"), auditResults
	}
	return nil, auditResults

}
