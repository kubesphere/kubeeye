package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	versionsv1alpha2 "github.com/kubesphere/kubeeye/clients/informers/externalversions/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/server/query"
	"github.com/kubesphere/kubeeye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	"strings"
)

type InspectRule struct {
	Clients *kube.KubernetesClient
	Ctx     context.Context
	Factory versionsv1alpha2.InspectRuleInformer
}

func NewInspectRule(ctx context.Context, clients *kube.KubernetesClient, f versionsv1alpha2.InspectRuleInformer) *InspectRule {
	return &InspectRule{
		Clients: clients,
		Ctx:     ctx,
		Factory: f,
	}
}

// ListInspectRule  godoc
// @Summary      Show an Inspect
// @Description  ListInspectRule
// @Tags         InspectRule
// @Accept       json
// @Produce      json
// @Param        sortBy query string false "sortBy=createTime"
// @Param        ascending query string false "ascending=true"
// @Param        limit query int false "limit=10"
// @Param        page query int false "page=1"
// @Param        labelSelector query string false "labelSelector=app=nginx"
// @Success      200 {array} v1alpha2.InspectRule
// @Router       /inspectrules [get]
func (i *InspectRule) ListInspectRule(g *gin.Context) {
	q := query.ParseQuery(g)

	parse, err := labels.Parse(q.LabelSelector)
	if err != nil {
		g.JSON(http.StatusInternalServerError, err)
		return
	}
	ret, err := i.Factory.Lister().List(parse)
	if err != nil {
		g.JSON(http.StatusInternalServerError, err)
		return
	}
	data := q.GetPageData(ret, i.compare, i.filter)

	g.JSON(http.StatusOK, data)
}

// GetInspectRule  godoc
// @Summary      Show an Inspect
// @Description  GetInspectRule
// @Tags         InspectRule
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Success      200 {object} v1alpha2.InspectRule
// @Router       /inspectrules/{name} [get]
func (i *InspectRule) GetInspectRule(gin *gin.Context) {
	name := gin.Param("name")
	rule, err := i.Factory.Lister().Get(name)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, rule)
}

// CreateInspectRule  godoc
// @Summary      Show an Inspect
// @Description  CreateInspectRule
// @Tags         InspectRule
// @Accept       json
// @Produce      json
// @Param		 InspectRule body	v1alpha2.InspectRule true	"Add InspectRule"
// @Success      200 {object} v1alpha2.InspectRule
// @Router       /inspectrules [post]
func (i *InspectRule) CreateInspectRule(gin *gin.Context) {
	var crateRule v1alpha2.InspectRule
	err := GetRequestBody(gin, &crateRule)
	if err != nil {
		Errs := NewErrors("bind data error", "InspectRule")
		gin.JSON(http.StatusInternalServerError, Errs)
		return
	}

	task, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().Create(i.Ctx, &crateRule, metav1.CreateOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, task)
}

// DeleteInspectRule  godoc
// @Summary      Show an Inspect
// @Description  DeleteInspectRule
// @Tags         InspectRule
// @Accept       json
// @Produce      json
// @Param        name path string true "name"
// @Success      200 {object} v1alpha2.InspectRule
// @Router       /inspectrules/{name} [delete]
func (i *InspectRule) DeleteInspectRule(gin *gin.Context) {
	name := gin.Param("name")
	err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().Delete(i.Ctx, name, metav1.DeleteOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.String(http.StatusOK, "success")
}

// UpdateInspectRule  godoc
// @Summary      Show an Inspect
// @Description  UpdateInspectRule
// @Tags         InspectRule
// @Accept       json
// @Produce      json
// @Param		 v1alpha2.InspectRule body	v1alpha2.InspectRule true	"Add InspectRule"
// @Success      200 {object} v1alpha2.InspectRule
// @Router       /inspectrules [put]
func (i *InspectRule) UpdateInspectRule(gin *gin.Context) {
	var updateRule v1alpha2.InspectRule
	err := GetRequestBody(gin, &updateRule)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	rule, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().Update(i.Ctx, &updateRule, metav1.UpdateOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, rule)
}

func (i *InspectRule) Validate(gin *gin.Context) {
	var crateRule v1alpha2.InspectRule

	err := GetRequestBody(gin, &crateRule)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		gin.Abort()
		return
	}

	_, ok := crateRule.GetLabels()[constant.LabelRuleGroup]
	if !ok {
		ResultErr := NewErrors(fmt.Sprintf("inspect rule must have label %s", constant.LabelRuleGroup), "InspectRule")
		gin.JSON(http.StatusInternalServerError, ResultErr)
		gin.Abort()
		return
	}

}

func (i *InspectRule) compare(a, b map[string]interface{}, orderBy string) bool {
	left := utils.MapToStruct[v1alpha2.InspectRule](a)
	right := utils.MapToStruct[v1alpha2.InspectRule](b)

	switch orderBy {
	case query.CreateTime:
		return left[0].CreationTimestamp.Before(&right[0].CreationTimestamp)
	case query.Name:
		return strings.Compare(left[0].Name, right[0].Name) < 0
	default:
		return false
	}

}

func (i *InspectRule) filter(data map[string]interface{}, f *query.Filter) bool {
	result := utils.MapToStruct[v1alpha2.InspectRule](data)[0]
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
