package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	versionsv1alpha2 "github.com/kubesphere/kubeeye/clients/informers/externalversions/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/server/query"
	"github.com/kubesphere/kubeeye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	"strings"
)

type InspectTask struct {
	Clients *kube.KubernetesClient
	Ctx     context.Context
	Factory versionsv1alpha2.InspectTaskInformer
}

func NewInspectTask(ctx context.Context, clients *kube.KubernetesClient, f versionsv1alpha2.InspectTaskInformer) *InspectTask {
	return &InspectTask{
		Clients: clients,
		Ctx:     ctx,
		Factory: f,
	}
}

// ListInspectTask  godoc
// @Summary      Show an Inspect
// @Description  ListInspectTask
// @Tags         InspectTask
// @Accept       json
// @Produce      json
// @Param        sortBy query string false "sortBy=createTime"
// @Param        ascending query string false "ascending=true"
// @Param        limit query int false "limit=10"
// @Param        page query int false "page=1"
// @Param        labelSelector query string false "labelSelector=app=nginx"
// @Success      200 {array} v1alpha2.InspectTask
// @Router       /inspecttasks [get]
func (i *InspectTask) ListInspectTask(gin *gin.Context) {
	q := query.ParseQuery(gin)
	parse, err := labels.Parse(q.LabelSelector)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	ret, err := i.Factory.Lister().List(parse)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	data := q.GetPageData(ret, i.compare, i.filter)
	gin.JSON(http.StatusOK, data)
}

// GetInspectTask  godoc
// @Summary      Show an Inspect
// @Description  ListInspectTask
// @Tags         InspectTask
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Success      200 {object} v1alpha2.InspectTask
// @Router       /inspecttasks/{name} [get]
func (i *InspectTask) GetInspectTask(gin *gin.Context) {
	name := gin.Param("name")
	task, err := i.Factory.Lister().Get(name)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, task)
}

// DeleteInspectTask  godoc
// @Summary      Show an Inspect
// @Description  DeleteInspectTask
// @Tags         InspectTask
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Success      200 {object} v1alpha2.InspectTask
// @Router       /inspecttasks/{name} [delete]
func (i *InspectTask) DeleteInspectTask(gin *gin.Context) {
	name := gin.Param("name")
	err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks().Delete(i.Ctx, name, metav1.DeleteOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.String(http.StatusOK, "success")
}

func (i *InspectTask) compare(a, b map[string]interface{}, orderBy string) bool {
	left := utils.MapToStruct[v1alpha2.InspectTask](a)[0]
	right := utils.MapToStruct[v1alpha2.InspectTask](b)[0]

	switch orderBy {
	case query.CreateTime:
		return left.CreationTimestamp.Before(&right.CreationTimestamp)
	case query.StartTimestamp:
		return left.Status.StartTimestamp.Before(right.Status.StartTimestamp)
	case query.Status:
		return strings.Compare(string(left.Status.Status), string(right.Status.Status)) < 0
	case query.InspectPolicy:
		return strings.Compare(string(left.Spec.InspectPolicy), string(right.Spec.InspectPolicy)) < 0
	case query.Duration:
		return strings.Compare(left.Status.Duration, right.Status.Duration) < 0
	default:
		return false
	}

}

func (i *InspectTask) filter(data map[string]interface{}, f *query.Filter) bool {
	result := utils.MapToStruct[v1alpha2.InspectTask](data)[0]
	for k, v := range *f {
		switch k {
		case query.Status:
			return result.Status.Status == v1alpha2.Phase(v)

		case query.InspectPolicy:
			return result.Spec.InspectPolicy == v1alpha2.Policy(v)
		default:
			return false
		}
	}
	return false
}
