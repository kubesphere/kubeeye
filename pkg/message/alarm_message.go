package message

import (
	"bytes"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/conf"

	"io"
	"k8s.io/klog/v2"
	"net/http"
)

type AlarmMessageHandler struct {
	// 可以添加处理器需要的属性
	RequestUrl string
}

func (h *AlarmMessageHandler) HandleMessageEvent(event *conf.MessageEvent) {
	// 执行消息发送操作
	// 例如，发送消息给目标

	fmt.Printf("Message sent to %s by %s: %s\n", event.Target, event.Sender, event.Content)
	resp, err := http.Post(h.RequestUrl, "application/json", bytes.NewReader(event.Content))
	if err != nil {
		klog.Error(err)
		return
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return
	}
	klog.Info(string(all))
}
