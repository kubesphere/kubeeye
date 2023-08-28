package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/controllers"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/server/query"
	"github.com/kubesphere/kubeeye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
)

type InspectTask struct {
	Clients *kube.KubernetesClient
	Ctx     context.Context
}

func NewInspectTask(ctx context.Context, clients *kube.KubernetesClient) *InspectTask {
	return &InspectTask{
		Clients: clients,
		Ctx:     ctx,
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
	list, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks().List(i.Ctx, metav1.ListOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	data := q.GetPageData(list.Items, i.compare, i.filter)
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
	task, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks().Get(i.Ctx, name, metav1.GetOptions{})
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
// @Param		 v1alpha2.InspectTask body	v1alpha2.InspectTask true	"delete InspectTask"
// @Success      200 {object} v1alpha2.InspectTask
// @Router       /inspectplans [delete]
func (i *InspectTask) DeleteInspectTask(gin *gin.Context) {
	var deleteTask v1alpha2.InspectTask
	err := gin.Bind(&deleteTask)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	err = i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks().Delete(i.Ctx, deleteTask.Name, metav1.DeleteOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, deleteTask)
}

func (i *InspectTask) compare(a, b map[string]interface{}, orderBy string) bool {
	left := utils.MapToStruct[v1alpha2.InspectTask](a)
	right := utils.MapToStruct[v1alpha2.InspectTask](b)

	switch orderBy {
	case query.CreateTime:
		return left[0].CreationTimestamp.Before(&right[0].CreationTimestamp)
	case query.Phase:
		return true
	case query.InspectPolicy:
		return strings.Compare(string(left[0].Spec.InspectPolicy), string(right[0].Spec.InspectPolicy)) < 0
	case query.Duration:
		return strings.Compare(left[0].Status.Duration, right[0].Status.Duration) < 0
	default:
		return false
	}

}

func (i *InspectTask) filter(data map[string]interface{}, f *query.Filter) bool {
	result := utils.MapToStruct[v1alpha2.InspectTask](data)[0]
	for k, v := range *f {
		switch k {
		case query.Phase:
			return controllers.GetStatus(&result) == v1alpha2.Phase(v)

		case query.InspectPolicy:
			return result.Spec.InspectPolicy == v1alpha2.Policy(v)
		default:
			return false
		}
	}
	return false
}
