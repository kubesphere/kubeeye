package validator

import (
	"context"
	"github.com/pkg/errors"
	"kubeye/pkg/config"
	"kubeye/pkg/kube"
	"time"
)

func ValidatePods(ctx context.Context, conf *config.Configuration, kubeResource *kube.ResourceProvider) ([]PodResult, error) {
	podToAudit := kubeResource.Controllers

	results := []PodResult{}

	for _, pod := range podToAudit {
		result, err := ValidatePod(ctx, conf, pod)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get result")
		}

		if len(result.ContainerResults[0].Results) == 0 || result.ContainerResults == nil {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func ValidatePod(ctx context.Context, c *config.Configuration, pod kube.GenericWorkload) (PodResult, error) {
	_, err := applyPodSchemaChecks(c, pod)
	if err != nil {
		return PodResult{}, err
	}
	pRes := PodResult{
		//Results:          podResults,
		ContainerResults: []ContainerResult{},
	}

	pRes.ContainerResults, err = ValidateAllContainers(ctx, c, pod)
	if err != nil {
		return pRes, err
	}

	result := PodResult{
		CreatedTime:      time.Now().Format(time.RFC3339),
		Kind:             pod.Kind,
		Name:             pod.ObjectMeta.GetName(),
		Namespace:        pod.ObjectMeta.GetNamespace(),
		ContainerResults: pRes.ContainerResults,
	}
	return result, nil

}
