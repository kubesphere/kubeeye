package pkg

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/armosec/opa-utils/reporthandling"
	"github.com/pkg/errors"
)

func KubescapeAudit() (result []reporthandling.FrameworkReport, err error) {
	log.Println("start KubeScape inspect")
	cmd := exec.Command("kubescape", "scan", "-e", "kube-system", "-f", "json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if strings.Contains(line, "controlReports") && strings.Contains(line, "services") {
			log.Println("decode KubeScape result")
			err := json.NewDecoder(strings.NewReader(line)).Decode(&result)
			if err != nil {
				return nil, errors.Wrap(err, "decode KubeScape result failed")
			}
		}
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "get KubeScape results failed")
		}

	}

	if err := cmd.Wait(); err != nil {
		return nil, errors.Wrap(err, "the command KubeScape exec failed")
	}
	log.Println("KubeScape inspect finished")
	return result, nil
}
