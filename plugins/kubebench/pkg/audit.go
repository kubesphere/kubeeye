package pkg

import (
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func KubeBenchAudit() (result *KubeBenchResults, err error) {
	log.Println("start KubeBench inspect")
	// exec KubeBench inspect, put the result into cmd
	cmd := exec.Command("kube-bench", "--json")
	// get the response from cmd.Output(), it contains the result of KubeBench inspect
	response, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute KubeBench")
	}

	log.Println("decode KubeBench inspect result")
	// decode the result of KubeBench inspect, put the result into allControls
	decoder := json.NewDecoder(strings.NewReader(string(response)))
	err = decoder.Decode(&result)
	if err == io.EOF {
		return nil, errors.Wrap(err, "the result of KubeBench are empty")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode the result of KubeBench")
	}

	log.Println("KubeBench inspect finished")
	return result, nil
}
