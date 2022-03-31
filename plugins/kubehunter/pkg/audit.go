package pkg

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

func KubehunterAudit(logs logr.Logger) (err error, kubehunterResults KubeHunterResults) {
	logs.Info("start kubehunter audit")
	cmd := exec.Command("kube-hunter", "--pod", "--report", "json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "the command kube-hunter create pipe failed"), kubehunterResults
	}
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "the command kube-hunter start failed"), kubehunterResults
	}
	reader := bufio.NewReader(stdout)

	for {
		line, err := reader.ReadString('\n')
		if strings.Contains(line, "nodes") && strings.Contains(line, "services") {
			logs.Info("decode result")
			err := json.NewDecoder(strings.NewReader(line)).Decode(&kubehunterResults)
			if err != nil {
				return errors.Wrap(err, "decode result failed"), kubehunterResults
			}
		}
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "get results failed"), kubehunterResults
		}

	}

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "the command kube-hunter exec failed"), kubehunterResults
	}
	return nil, kubehunterResults
}
