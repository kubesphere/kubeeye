package pkg

import (
	"github.com/aquasecurity/kube-bench/check"
)

var KubeBenchResult KubeBenchResults
var KBResult = &KubeBenchResult

type KubeBenchResults struct {
	Controls []check.Controls
}
