package message

import "github.com/kubesphere/kubeeye/pkg/conf"

type EventDispatcher struct {
	handlers conf.EventHandler
}

func RegisterHandler(handler conf.EventHandler) *EventDispatcher {
	return &EventDispatcher{
		handlers: handler,
	}
}

func (d *EventDispatcher) DispatchMessageEvent(event *conf.MessageEvent) {
	d.handlers.HandleMessageEvent(event)
}
