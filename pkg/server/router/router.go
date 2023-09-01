package router

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/clients/informers/externalversions/kubeeye"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/server/api"
	"github.com/kubesphere/kubeeye/pkg/template"
)

const groupPath = "/kapis/kubeeye.kubesphere.io/v1alpha2"

type Router struct {
	Engine  *gin.Engine
	Clients *kube.KubernetesClient
	Ctx     context.Context
}

func RegisterRouter(ctx context.Context, r *gin.Engine, clients *kube.KubernetesClient, factory kubeeye.Interface) {
	result := api.NewInspectResult(ctx, clients, factory.V1alpha2().InspectResults())
	task := api.NewInspectTask(ctx, clients, factory.V1alpha2().InspectTasks())
	plan := api.NewInspectPlan(ctx, clients, factory.V1alpha2().InspectPlans())
	rule := api.NewInspectRule(ctx, clients, factory.V1alpha2().InspectRules())
	htmlTemplate, err := template.GetInspectResultHtmlTemplate()
	if err == nil {
		r.SetHTMLTemplate(htmlTemplate)
	}

	v1alpha1 := r.Group(groupPath)
	{

		v1alpha1.GET("/inspectresults", result.ListInspectResult)
		v1alpha1.GET("/inspectresults/:name", result.GetInspectResult)

		v1alpha1.GET("/inspecttasks", task.ListInspectTask)
		v1alpha1.GET("/inspecttasks/:name", task.GetInspectTask)
		v1alpha1.DELETE("/inspecttasks/:name", task.DeleteInspectTask)

		v1alpha1.GET("/inspectplans", plan.ListInspectPlan)
		v1alpha1.GET("/inspectplans/:name", plan.GetInspectPlan)
		v1alpha1.POST("/inspectplans", plan.CreateInspectPlan)
		v1alpha1.DELETE("/inspectplans/:name", plan.DeleteInspectPlan)
		v1alpha1.PUT("/inspectplans", plan.UpdateInspectPlan)
		v1alpha1.PATCH("/inspectplans/:name/status", plan.PatchInspectPlanStatus)
		v1alpha1.PATCH("/inspectplans/:name", plan.PatchInspectPlan)

		v1alpha1.GET("/inspectrules", rule.ListInspectRule)
		v1alpha1.GET("/inspectrules/:name", rule.GetInspectRule)
		v1alpha1.POST("/inspectrules", rule.CreateInspectRule)
		v1alpha1.DELETE("/inspectrules", rule.DeleteInspectRule)
		v1alpha1.PUT("/inspectrules", rule.UpdateInspectRule)
	}

}
