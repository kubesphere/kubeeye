package v1

import (
	"github.com/emicklei/go-restful"
)

type handler struct {
}

func newHandler() *handler {
	return &handler{}
}

func (h handler) OverView(request *restful.Request, response *restful.Response) {
	// logical process
	//result := dashboard.GetInfo(context.Background())
	//response.WriteAsJson(result)
}
