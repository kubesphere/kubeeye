package kubeeye

import (
	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/kube"
)

func setClusterInfo(k8SResource kube.K8SResource) (ClusterInfo kubeeyev1alpha1.ClusterInfo) {
	ClusterInfo.ClusterVersion = k8SResource.ServerVersion
	ClusterInfo.NodesCount = k8SResource.NodesCount
	ClusterInfo.NamespacesCount = k8SResource.NameSpacesCount
	ClusterInfo.NamespacesList = k8SResource.NameSpacesList
	ClusterInfo.WorkloadsCount = k8SResource.WorkloadsCount
	return ClusterInfo
}

func formatResults(receiver <-chan []kubeeyev1alpha1.AuditResults) (formattedResults []kubeeyev1alpha1.AuditResults) {
	var formattedResult kubeeyev1alpha1.AuditResults
	fmAuditResults := make(map[string][]kubeeyev1alpha1.ResultInfos)

	for results := range receiver {
		for _, result := range results {
			fmAuditResults[result.NameSpace] = append(fmAuditResults[result.NameSpace], result.ResultInfos...)
		}
	}

	for nm, ar := range fmAuditResults {
		formattedResult.ResultInfos = ar
		formattedResult.NameSpace = nm
		formattedResults = append(formattedResults, formattedResult)
	}

	return formattedResults
}

func CalculateScore(fmResultss []kubeeyev1alpha1.AuditResults, k8sResources kube.K8SResource) (scoreInfo kubeeyev1alpha1.ScoreInfo) {
	var countDanger int
	var countWarning int
	var countIgnore int

	for _, fmResult := range fmResultss {
		for _, resultInfo := range fmResult.ResultInfos {
			for _, item := range resultInfo.ResultItems {
				if item.Level == "warning" {
					countWarning++
				} else if item.Level == "danger" {
					countDanger++
				} else if item.Level == "ignore" {
					countIgnore++
				}
			}
		}
	}

	total := k8sResources.WorkloadsCount*20 + (len(k8sResources.Roles.Items)+len(k8sResources.ClusterRoles.Items))*3 + len(k8sResources.Events.Items) + len(k8sResources.Nodes.Items) + 1
	countSuccess := total - countDanger - countWarning - countIgnore
	totalWeight := countSuccess*2 + countDanger*2 + countWarning
	scoreInfo.Score = countSuccess * 2 * 100 / totalWeight
	scoreInfo.Total = total
	scoreInfo.Dangerous = countDanger
	scoreInfo.Warning = countWarning
	scoreInfo.Ignore = countIgnore
	scoreInfo.Passing = countSuccess

	return scoreInfo
}
