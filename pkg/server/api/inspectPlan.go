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
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"strings"
)

type InspectPlan struct {
	Clients *kube.KubernetesClient
	Ctx     context.Context
	Factory versionsv1alpha2.InspectPlanInformer
}

func NewInspectPlan(ctx context.Context, clients *kube.KubernetesClient, f versionsv1alpha2.InspectPlanInformer) *InspectPlan {
	return &InspectPlan{
		Clients: clients,
		Ctx:     ctx,
		Factory: f,
	}
}

// ListInspectPlan  godoc
// @Summary      Show an Inspect
// @Description  ListInspectPlan
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param        sortBy query string false "sortBy=createTime"
// @Param        ascending query string false "ascending=true"
// @Param        limit query int false "limit=10"
// @Param        page query int false "page=1"
// @Param        labelSelector query string false "labelSelector=app=nginx"
// @Success      200 {array} v1alpha2.InspectPlan
// @Router       /inspectplans [get]
func (i *InspectPlan) ListInspectPlan(gin *gin.Context) {
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
	plan, err := i.Factory.Lister().Get(name)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, plan)
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
// @Param        name path string true "name"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectplans/{name} [delete]
func (i *InspectPlan) DeleteInspectPlan(gin *gin.Context) {

	name := gin.Param("name")
	err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Delete(i.Ctx, name, metav1.DeleteOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.String(http.StatusOK, "success")
}

// UpdateInspectPlan  godoc
// @Summary      Show an Inspect
// @Description  UpdateInspectRule
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param		 v1alpha2.InspectPlan body	v1alpha2.InspectPlan true	"update InspectPlan"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectplans [put]
func (i *InspectPlan) UpdateInspectPlan(g *gin.Context) {
	var updatePlan v1alpha2.InspectPlan
	err := GetRequestBody(g, &updatePlan)
	if err != nil {
		g.JSON(http.StatusInternalServerError, NewErrors(err.Error(), "InspectPlan"))
		return
	}
	rule, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Update(i.Ctx, &updatePlan, metav1.UpdateOptions{})
	if err != nil {
		g.JSON(http.StatusInternalServerError, err)
		return
	}
	g.JSON(http.StatusOK, rule)
}

// PatchInspectPlan   godoc
// @Summary      Show an Inspect
// @Description  PatchInspectPlan
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Param		 v1alpha2.InspectPlan body	v1alpha2.InspectPlan true	"patch InspectPlan"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectplans/{name} [patch]
func (i *InspectPlan) PatchInspectPlan(g *gin.Context) {
	name := g.Param("name")
	data, err := g.GetRawData()
	if err != nil {
		g.JSON(http.StatusInternalServerError, NewErrors(err.Error(), "InspectPlan"))
		return
	}
	result, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Patch(i.Ctx, name, types.MergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		g.JSON(http.StatusInternalServerError, err)
		return
	}
	g.JSON(http.StatusOK, result)
}

// PatchInspectPlanStatus   godoc
// @Summary      Show an Inspect
// @Description  PatchInspectPlanStatus
// @Tags         InspectPlan
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Param		 status body v1alpha2.InspectPlanStatus true "{status:{lastTaskStatus:‘’}}"
// @Success      200 {object} v1alpha2.InspectPlan
// @Router       /inspectplans/{name}/status [patch]
func (i *InspectPlan) PatchInspectPlanStatus(g *gin.Context) {
	name := g.Param("name")
	plan, err := i.Factory.Lister().Get(name)
	if err != nil {
		g.JSON(http.StatusInternalServerError, NewErrors(err.Error(), "InspectPlan"))
	}
	if plan.Status.LastTaskStatus.IsPending() || plan.Status.LastTaskStatus.IsRunning() {
		g.JSON(http.StatusInternalServerError, NewErrors("Please do not repeat the execution", "InspectPlan"))
		return
	}

	data, err := g.GetRawData()
	if err != nil {
		g.JSON(http.StatusInternalServerError, NewErrors(err.Error(), "InspectPlan"))
		return
	}

	result, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectPlans().Patch(i.Ctx, name, types.MergePatchType, data, metav1.PatchOptions{}, "status")
	if err != nil {
		g.JSON(http.StatusInternalServerError, err)
		return

	}
	g.JSON(http.StatusOK, result)
}

func (i *InspectPlan) compare(a, b map[string]interface{}, orderBy string) bool {
	left := utils.MapToStruct[v1alpha2.InspectPlan](a)
	right := utils.MapToStruct[v1alpha2.InspectPlan](b)

	switch orderBy {
	case query.CreateTime:
		return left[0].CreationTimestamp.Before(&right[0].CreationTimestamp)
	case query.LastTaskStatus:
		return strings.Compare(string(left[0].Status.LastTaskStatus), string(right[0].Status.LastTaskStatus)) < 0
	case query.LastTaskStartTime:
		return left[0].Status.LastTaskStartTime.Before(right[0].Status.LastTaskStartTime)
	case query.InspectPolicy:
		return true
	default:
		return false
	}

}

func (i *InspectPlan) filter(data map[string]interface{}, f *query.Filter) bool {
	result := utils.MapToStruct[v1alpha2.InspectPlan](data)[0]
	isTag := false
	for k, v := range *f {
		switch k {
		case query.Suspend:
			isTag = result.Spec.Suspend == utils.StringToBool(v)
		case query.LastTaskStatus:
			isTag = result.Status.LastTaskStatus == v1alpha2.Phase(v)
		case query.Strategy:
			if v1alpha2.Policy(v) == v1alpha2.CyclePolicy {
				isTag = result.Spec.Once == nil && result.Spec.Schedule != nil
			} else {
				isTag = result.Spec.Once != nil
			}
		case query.InspectType:
			if v1alpha2.Policy(v) == v1alpha2.InspectTypeInstant {
				isTag = result.Spec.Schedule == nil && result.Spec.Once == nil
			} else {
				isTag = result.Spec.Schedule != nil || result.Spec.Once != nil
			}
		case query.Name:
			isTag = strings.Contains(result.Name, v)
		}
		if !isTag {
			return false
		}
	}
	return true
}
