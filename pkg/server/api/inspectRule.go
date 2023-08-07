package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/cmd/apiserver/options"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sort"
)

type InspectRule struct {
	Clients *kube.KubernetesClient
	Ctx     context.Context
}

func NewInspectRule(ctx context.Context, clients *kube.KubernetesClient) *InspectRule {
	return &InspectRule{
		Clients: clients,
		Ctx:     ctx,
	}
}

// ListInspectRule  godoc
// @Summary      Show an Inspect
// @Description  ListInspectRule
// @Tags         InspectRule
// @Accept       json
// @Produce      json
// @Param        orderBy query string false "orderBy"
// @Param        ascending query string false "ascending"
// @Success      200 {array} v1alpha2.InspectRule
// @Router       /inspectrules [get]
func (i *InspectRule) ListInspectRule(gin *gin.Context) {
	list, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().List(i.Ctx, metav1.ListOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	InspectRuleSortBy(list.Items, gin)
	gin.JSON(http.StatusOK, list.Items)
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
	task, err := i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().Get(i.Ctx, name, metav1.GetOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, task)
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
	err := gin.Bind(&crateRule)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
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
// @Param		 v1alpha2.InspectRule body	v1alpha2.InspectRule true	"Add InspectRule"
// @Success      200 {object} v1alpha2.InspectRule
// @Router       /inspectrules [delete]
func (i *InspectRule) DeleteInspectRule(gin *gin.Context) {
	var deleteRule v1alpha2.InspectRule
	err := gin.Bind(&deleteRule)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	err = i.Clients.VersionClientSet.KubeeyeV1alpha2().InspectRules().Delete(i.Ctx, deleteRule.Name, metav1.DeleteOptions{})
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		return
	}
	gin.JSON(http.StatusOK, deleteRule)
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
	err := gin.Bind(&updateRule)
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

func InspectRuleSortBy(tasks []v1alpha2.InspectRule, gin *gin.Context) {
	orderBy := gin.Query(options.OrderBy)
	asc := gin.Query(options.Ascending)
	sort.Slice(tasks, func(i, j int) bool {
		if asc == "true" {
			i, j = j, i
		}
		return ObjectMetaCompare(tasks[i].ObjectMeta, tasks[j].ObjectMeta, orderBy)
	})

}

func (i *InspectRule) Validate(gin *gin.Context) {
	var crateRule v1alpha2.InspectRule
	err := gin.Bind(&crateRule)
	if err != nil {
		gin.JSON(http.StatusInternalServerError, err)
		gin.Abort()
		return
	}
	_, ok := crateRule.GetLabels()[constant.LabelInspectRuleGroup]
	if !ok {
		gin.String(http.StatusInternalServerError, fmt.Sprintf("inspect rule must have label %s", constant.LabelInspectRuleGroup))
		gin.Abort()
		return
	}
	gin.Next()

}
