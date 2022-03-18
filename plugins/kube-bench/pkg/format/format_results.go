package format

import (
	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/api/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/audit"
)

func ResultsFormat(KubeBenchAudit audit.KubeBenchResponse) (auditResults []kubeeyev1alpha1.AuditResults) {
	for _, controls := range KubeBenchAudit.Controls {
		var auditResult kubeeyev1alpha1.AuditResults
		var resultInfos kubeeyev1alpha1.ResultInfos

		resultInfos.ResourceType = string(controls.Type)
		for _, group := range controls.Groups {
			for _, checks := range group.Checks {
				var resourceInfos kubeeyev1alpha1.ResourceInfos
				var resultItems kubeeyev1alpha1.ResultItems

				if checks.State != "PASS" {
					resourceInfos.Name = checks.Text
					resultItems.Level = "warning"
					resultItems.Message = checks.Remediation
					resultItems.Reason = checks.Reason
					resourceInfos.ResultItems = append(resourceInfos.ResultItems, resultItems)
					resultInfos.ResourceInfos = resourceInfos
					auditResult.ResultInfos = append(auditResult.ResultInfos, resultInfos)
				}
			}
		}

		auditResults = append(auditResults, auditResult)
	}
	return auditResults
}
