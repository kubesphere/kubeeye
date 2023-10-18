package conf

import (
	"time"
)

type MessageEvent struct {
	Content   string
	Target    string
	Sender    string
	Timestamp time.Time
}

type EventHandler interface {
	HandleMessageEvent(event *MessageEvent)
}
