package expend

import (
	"context"
	"fmt"
	"strings"
)

func UninstallExpendPackages(ctx context.Context, kubeconfig string, expendPackages string) error {
	packagesname := strings.Split(expendPackages, ",")
	for _, packagename := range packagesname {
		switch packagename {
		case "npd", "node-problem-detector", "NPD":
			UninstallNPD(ctx, kubeconfig)
		default:
			err := fmt.Errorf("unable to install the package: %s", packagename)
			return err
		}
	}
	return nil
}

func UninstallNPD(ctx context.Context, kubeconfig string) {
	var npd Expends
	npdStuck := NPD{kubeconfig: kubeconfig, ctx: ctx}
	npd = npdStuck
	npd.uninstall()
}
