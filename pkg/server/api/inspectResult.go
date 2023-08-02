package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/cmd/apiserver/options"
	"github.com/kubesphere/kubeeye/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"sort"
)

type InspectResult struct {
	Clients *kube.KubernetesClient
	Ctx     context.Context
}

func NewInspectResult(ctx context.Context, clients *kube.KubernetesClient) *InspectResult {
	return &InspectResult{
		Clients: clients,
		Ctx:     ctx,
	}
}

// ListInspectResult godoc
// @Summary      Show an Inspect
// @Description  get ListInspectResult
// @Tags         InspectResult
// @Accept       json
// @Produce      json
// @Param        orderBy query string false "orderBy"
// @Param        ascending query string false "ascending"
// @Success      200 {array} v1alpha2.InspectResult
// @Router       /inspectresults [get]
func (o *InspectResult) ListInspectResult(gin *gin.Context) {
	list, err := o.Clients.VersionClientSet.KubeeyeV1alpha2().InspectResults().List(o.Ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error(err)
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	InspectResultSortBy(list.Items, gin)
	gin.JSON(http.StatusOK, list.Items)
}

// GetInspectResult godoc
// @Summary      Show an Inspect
// @Description  GetInspectResult
// @Tags         InspectResult
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Success      200 {object} v1alpha2.InspectResult
// @Router       /inspectresults/{name} [get]
func (o *InspectResult) GetInspectResult(gin *gin.Context) {
	name := gin.Param("name")
	list, err := o.Clients.VersionClientSet.KubeeyeV1alpha2().InspectResults().Get(o.Ctx, name, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		gin.JSON(http.StatusInternalServerError, err)
		return
	}

	gin.JSON(http.StatusOK, list)

}

func InspectResultSortBy(result []v1alpha2.InspectResult, gin *gin.Context) {
	orderBy := gin.Query(options.OrderBy)
	asc := gin.Query(options.Ascending)
	sort.Slice(result, func(i, j int) bool {
		if asc == "true" {
			i, j = j, i
		}
		return ObjectMetaCompare(result[i].ObjectMeta, result[j].ObjectMeta, orderBy)
	})

}

func ObjectMetaCompare(a, b metav1.ObjectMeta, orderBy string) bool {
	switch orderBy {
	case options.CreateTime:
		return a.CreationTimestamp.Before(&b.CreationTimestamp)
	}
	return false
}
