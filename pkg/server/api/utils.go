package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func GetRequestBody(g *gin.Context, obj any) error {
	data, err := g.GetRawData()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return err
	}
	g.Request.Body = io.NopCloser(bytes.NewBuffer(data))
	return nil
}

func NewErrors(msg string, kind string) errors.StatusError {
	return errors.StatusError{ErrStatus: metav1.Status{
		Status:  "Failure",
		Message: msg,
		Details: &metav1.StatusDetails{
			Group:  v1alpha2.SchemeGroupVersion.Group,
			Kind:   kind,
			Causes: nil,
		},
		Code: http.StatusInternalServerError,
	}}
}
