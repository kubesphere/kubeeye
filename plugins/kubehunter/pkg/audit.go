package pkg

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func KubeHunterAudit() (result *KubeHunterResults, err error) {
	log.Println("start kubehunter audit")
	cmd := exec.Command("kube-hunter", "--pod", "--report", "json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "the command kube-hunter create pipe failed")
	}
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "the command kube-hunter start failed")
	}
	reader := bufio.NewReader(stdout)

	for {
		line, err := reader.ReadString('\n')
		if strings.Contains(line, "nodes") && strings.Contains(line, "services") {
			log.Println("decode result")
			err := json.NewDecoder(strings.NewReader(line)).Decode(&result)
			if err != nil {
				return nil, errors.Wrap(err, "decode result failed")
			}
		}
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return result, errors.Wrap(err, "get results failed")
		}

	}

	if err := cmd.Wait(); err != nil {
		return result, errors.Wrap(err, "the command kube-hunter exec failed")
	}
	log.Println("KubeHunter audit finished")
	return result, nil
}
