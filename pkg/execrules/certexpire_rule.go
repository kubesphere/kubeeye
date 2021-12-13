package execrules

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/register"
	"k8s.io/apimachinery/pkg/util/duration"
	certutil "k8s.io/client-go/util/cert"
)

func init() {
	register.ExecRuleRegistry(CertExpireRule{})
}

type certificate struct {
	Name     string `yaml:"name" json:"name,omitempty"`
	Expires  string `yaml:"expires" json:"expires,omitempty"`
	Residual string `yaml:"residual" json:"residual,omitempty"`
}

type CertExpireRule struct{}

func (cer CertExpireRule) Exec() []kube.ResultReceiver {
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
	var output []kube.ResultReceiver
	var certExpiresOutput kube.ResultReceiver
	if len(certExpires) != 0 {
		for _, certExpire := range certExpires {
			if len(certExpire.Expires) != 0 {
				certExpiresOutput.Name = certExpire.Name
				certExpiresOutput.Type = "certExpire"
				certExpiresOutput.Message = append(certExpiresOutput.Message, certExpire.Expires, certExpire.Residual)

				output = append(output, certExpiresOutput)
			}
		}
	}
	return output
}
