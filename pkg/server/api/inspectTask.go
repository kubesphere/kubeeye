package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/cmd/apiserver/options"
	"github.com/kubesphere/kubeeye/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sort"
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
// @Param        orderBy query string false "orderBy"
// @Param        ascending query string false "ascending"
// @Success      200 {array} v1alpha2.InspectTask
// @Router       /inspecttasks [get]
func (i *InspectTask) ListInspectTask(gin *gin.Context) {
	list, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks().List(i.Ctx, metav1.ListOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	InspectTaskSortBy(list.Items, gin)
	gin.JSON(http.StatusOK, list.Items)
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

func InspectTaskSortBy(tasks []v1alpha2.InspectTask, gin *gin.Context) {
	orderBy := gin.Query(options.OrderBy)
	asc := gin.Query(options.Ascending)
	sort.Slice(tasks, func(i, j int) bool {
		if asc == "true" {
			i, j = j, i
		}
		return ObjectMetaCompare(tasks[i].ObjectMeta, tasks[j].ObjectMeta, orderBy)
	})

}
