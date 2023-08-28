package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/output"
	"github.com/kubesphere/kubeeye/pkg/server/query"
	"github.com/kubesphere/kubeeye/pkg/template"
	"github.com/kubesphere/kubeeye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
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
// @Param        sortBy query string false "sortBy=createTime"
// @Param        ascending query string false "ascending=true"
// @Param        limit query int false "limit=10"
// @Param        page query int false "page=1"
// @Param        labelSelector query string false "labelSelector=app=nginx"
// @Success      200 {array} v1alpha2.InspectResult
// @Router       /inspectresults [get]
func (i *InspectResult) ListInspectResult(gin *gin.Context) {
	q := query.ParseQuery(gin)
	list, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectResults().List(i.Ctx, metav1.ListOptions{
		LabelSelector: q.LabelSelector,
	})
	if err != nil {
		klog.Error(err)
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	data := q.GetPageData(list.Items, i.compare, i.filter)

	gin.JSON(http.StatusOK, data)
}

// GetInspectResult godoc
// @Summary      Show an Inspect
// @Description  GetInspectResult
// @Tags         InspectResult
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Param        type query string false "type"
// @Success      200 {object} v1alpha2.InspectResult
// @Router       /inspectresults/{name} [get]
func (i *InspectResult) GetInspectResult(gin *gin.Context) {
	name := gin.Param("name")
	outType := gin.Query("type")
	switch outType {
	case "html":
		err, m := output.HtmlOut(gin.Param("name"))
		if err != nil {
			gin.JSON(http.StatusInternalServerError, err)
			return
		}
		gin.HTML(http.StatusOK, template.InspectResultTemplate, m)
	default:
		list, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectResults().Get(i.Ctx, name, metav1.GetOptions{})
		if err != nil {
			klog.Error(err)
			gin.JSON(http.StatusInternalServerError, err)
			return
		}
		gin.JSON(http.StatusOK, list)
	}

}

func (i *InspectResult) compare(a, b map[string]interface{}, orderBy string) bool {
	left := utils.MapToStruct[v1alpha2.InspectResult](a)
	right := utils.MapToStruct[v1alpha2.InspectResult](b)

	switch orderBy {
	case query.CreateTime:
		return left[0].CreationTimestamp.Before(&right[0].CreationTimestamp)
	}
	return false
}

func (i *InspectResult) filter(data map[string]interface{}, f *query.Filter) bool {
	result := utils.MapToStruct[v1alpha2.InspectResult](data)[0]
	for k, v := range *f {
		switch k {
		case query.Name:
			return strings.Contains(result.Name, v)
		default:
			return false
		}
	}
	return false
}
