package wpc

import (
	"sync"
	"time"
)

var msgPool = sync.Pool{New: func() any {
	return new(Message)
}}

type Message struct {
	MessageId       []byte
	Topic           string
	ProducerName    string
	Properties      map[string]string
	Payload         []byte
	PublishTime     time.Time
	EventTime       time.Time
	RedeliveryCount int
}

func (m *Message) ReleaseThis() {
	ReleaseMessage(m)
}

func AcquireMessage() *Message {
	return msgPool.Get().(*Message)
}

func ReleaseMessage(msg *Message) {
	msg.MessageId = nil
	msg.Topic = ""
	msg.ProducerName = ""
	msg.Properties = nil
	msg.Payload = nil
	msg.PublishTime = time.Time{}
	msg.EventTime = time.Time{}
	msg.RedeliveryCount = 0
	msgPool.Put(msg)
}
