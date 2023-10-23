package message

import (
	"context"
	"github.com/kubesphere/kubeeye/pkg/conf"
	"github.com/kubesphere/kubeeye/pkg/constant"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"net/smtp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EmailMessageHandler struct {
	// 可以添加处理器需要的属性
	*conf.EmailConfig
	client.Client
}

func NewEmailMessageOptions(event *conf.EmailConfig, c client.Client) *EmailMessageHandler {
	return &EmailMessageHandler{
		EmailConfig: event,
		Client:      c,
	}
}

func (e *EmailMessageHandler) HandleMessageEvent(event *conf.MessageEvent) {
	// 执行消息发送操作
	// 例如，发送消息给目标
	if e != nil {
		return
	}
	var secret corev1.Secret
	err := e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: constant.DefaultNamespace,
		Name:      e.SecretKey,
	}, &secret)
	u := secret.StringData["username"]
	p := secret.StringData["password"]
	auth := smtp.PlainAuth("", u, p, e.Address)
	err = smtp.SendMail(e.Address, auth, e.Fo, e.To, event.Content)
	if err != nil {
		klog.Error("send email failed, err: ", err)
	}
}
