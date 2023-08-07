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

type InspectPlan struct {
	Clients *kube.KubernetesClient
	Ctx     context.Context
}

func NewInspectPlan(ctx context.Context, clients *kube.KubernetesClient) *InspectPlan {
	return &InspectPlan{
		Clients: clients,
		Ctx:     ctx,
	}
}

// ListInspectPlan  godoc
// @Summary      Show an Inspect
// @Description  ListInspectPlan
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param        orderBy query string false "orderBy"
// @Param        ascending query string false "ascending"
// @Success      200 {array} v1alpha2.InspectPlan
// @Router       /inspectplans [get]
func (i *InspectPlan) ListInspectPlan(gin *gin.Context) {
	list, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().List(i.Ctx, metav1.ListOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	InspectPlanSortBy(list.Items, gin)
	gin.JSON(http.StatusOK, list.Items)
}

// GetInspectPlan  godoc
// @Summary      Show an Inspect
// @Description  GetInspectPlan
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectplans/{name} [get]
func (i *InspectPlan) GetInspectPlan(gin *gin.Context) {
	name := gin.Param("name")
	task, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Get(i.Ctx, name, metav1.GetOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, task)
}

// CreateInspectPlan  godoc
// @Summary      Show an Inspect
// @Description  CreateInspectPlan
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param		 InspectPlan body	v1alpha2.InspectPlan true	"Add InspectPlan"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectplans [post]
func (i *InspectPlan) CreateInspectPlan(gin *gin.Context) {
	var cratePlan v1alpha2.InspectPlan
	err := gin.Bind(&cratePlan)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}

	plan, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Create(i.Ctx, &cratePlan, metav1.CreateOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, plan)
}

// DeleteInspectPlan  godoc
// @Summary      Show an Inspect
// @Description  DeleteInspectRule
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param		 v1alpha2.InspectPlan body	v1alpha2.InspectPlan true	"delete InspectPlan"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectplans [delete]
func (i *InspectPlan) DeleteInspectPlan(gin *gin.Context) {
	var deletePlan v1alpha2.InspectPlan
	err := gin.Bind(&deletePlan)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	err = i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Delete(i.Ctx, deletePlan.Name, metav1.DeleteOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, deletePlan)
}

// UpdateInspectPlan  godoc
// @Summary      Show an Inspect
// @Description  UpdateInspectRule
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param		 v1alpha2.InspectPlan body	v1alpha2.InspectPlan true	"update InspectPlan"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectrules [put]
func (i *InspectPlan) UpdateInspectPlan(gin *gin.Context) {
	var updatePlan v1alpha2.InspectPlan
	err := gin.Bind(&updatePlan)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	rule, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Update(i.Ctx, &updatePlan, metav1.UpdateOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, rule)
}

func InspectPlanSortBy(tasks []v1alpha2.InspectPlan, gin *gin.Context) {
	orderBy := gin.Query(options.OrderBy)
	asc := gin.Query(options.Ascending)
	sort.Slice(tasks, func(i, j int) bool {
		if asc == "true" {
			i, j = j, i
		}
		return ObjectMetaCompare(tasks[i].ObjectMeta, tasks[j].ObjectMeta, orderBy)
	})

}
