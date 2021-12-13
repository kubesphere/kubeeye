package execrules

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/leonharetd/kubeeye/pkg/kube"
	"github.com/leonharetd/kubeeye/pkg/register"
	"k8s.io/apimachinery/pkg/util/duration"
	certutil "k8s.io/client-go/util/cert"
)

func init() {
	register.FuncRuleRegistry(CertExpireRule{})
}

type certificate struct {
	Name     string `yaml:"name" json:"name,omitempty"`
	Expires  string `yaml:"expires" json:"expires,omitempty"`
	Residual string `yaml:"residual" json:"residual,omitempty"`
}

type CertExpireRule struct{}

func (cer CertExpireRule) Exec() kube.ValidateResults {
	var certExpires []certificate
	cmd := fmt.Sprintf("cat /etc/kubernetes/pki/%s", "apiserver.crt")
	combinedoutput, _ := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	if combinedoutput != nil {
		certs, _ := certutil.ParseCertsPEM([]byte(combinedoutput))
		if len(certs) != 0 {
			certExpire := certificate{
				Name:     "kube-apiserver",
				Expires:  certs[0].NotAfter.Format("Jan 02, 2006 15:04 MST"),
				Residual: duration.ShortHumanDuration(time.Until(certs[0].NotAfter)),
			}
			if strings.Contains(certExpire.Residual, "d") {
				tmpTime, _ := strconv.Atoi(strings.TrimRight(certExpire.Residual, "d"))
				if tmpTime < 30 {
					certExpires = append(certExpires, certExpire)
				}
			} else {
				certExpires = append(certExpires, certExpire)
			}
		}
	}
	output := kube.ValidateResults{ValidateResults: make([]kube.ResultReceiver, 0)}
	var certExpiresOutput kube.ResultReceiver
	if len(certExpires) != 0 {
		for _, certExpire := range certExpires {
			if len(certExpire.Expires) != 0 {
				certExpiresOutput.Name = certExpire.Name
				certExpiresOutput.Type = "certExpire"
				certExpiresOutput.Message = append(certExpiresOutput.Message, certExpire.Expires, certExpire.Residual)
				output.ValidateResults = append(output.ValidateResults, certExpiresOutput)
			}
		}
	}
	return output
}
