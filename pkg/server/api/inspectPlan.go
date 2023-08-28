package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/server/query"
	"github.com/kubesphere/kubeeye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
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
// @Param        orderBy query string false "orderBy=createTime"
// @Param        ascending query string false "ascending=true"
// @Param        limit query int false "limit=10"
// @Param        page query int false "page=1"
// @Param        labelSelector query string false "labelSelector=app=nginx"
// @Success      200 {array} v1alpha2.InspectPlan
// @Router       /inspectplans [get]
func (i *InspectPlan) ListInspectPlan(gin *gin.Context) {
	q := query.ParseQuery(gin)
	list, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().List(i.Ctx, metav1.ListOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	data := q.GetPageData(list.Items, i.compare, i.filter)
	gin.JSON(http.StatusOK, data)
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
	err := GetRequestBody(gin, &cratePlan)

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
func (i *InspectPlan) UpdateInspectPlan(g *gin.Context) {
	var updatePlan v1alpha2.InspectPlan
	err := GetRequestBody(g, &updatePlan)
	if err != nil {
		g.JSON(http.StatusInternalServerError, err)
		return
	}
	rule, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Update(i.Ctx, &updatePlan, metav1.UpdateOptions{})
	if err != nil {
		g.JSON(http.StatusInternalServerError, err)
		return
	}
	g.JSON(http.StatusOK, rule)
}

func (i *InspectPlan) compare(a, b map[string]interface{}, orderBy string) bool {
	left := utils.MapToStruct[v1alpha2.InspectPlan](a)
	right := utils.MapToStruct[v1alpha2.InspectPlan](b)

	switch orderBy {
	case query.CreateTime:
		return left[0].CreationTimestamp.Before(&right[0].CreationTimestamp)
	case query.Phase:
		return strings.Compare(string(left[0].Status.LastTaskStatus), string(right[0].Status.LastTaskStatus)) < 0
	case query.InspectPolicy:
		return true
	default:
		return false
	}

}

func (i *InspectPlan) filter(data map[string]interface{}, f *query.Filter) bool {
	result := utils.MapToStruct[v1alpha2.InspectPlan](data)[0]
	for k, v := range *f {
		switch k {
		case query.Suspend:
			return result.Spec.Suspend == utils.StringToBool(v)
		case query.Phase:
			return result.Status.LastTaskStatus == v1alpha2.Phase(v)
		case query.InspectPolicy:
			if v1alpha2.Policy(v) == v1alpha2.InstantPolicy {
				return result.Spec.Schedule == nil
			}
			return result.Spec.Schedule != nil
		case query.Name:
			return strings.Contains(result.Name, v)
		default:
			return false
		}
	}
	return false
}
