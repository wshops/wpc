package wpc

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gookit/slog"
	"github.com/wshops/wpc/workerpool"
	"sync"
	"time"
)

// https://github.com/chinagocoder/go-queue/blob/main/pq/consumer.go
// https://github.com/AlexCuse/watermill-pulsar/blob/main/pkg/pulsar/subscriber.go
// https://github.com/utain/poc-realtime-go/blob/main/withpulsar/withpulsar.go

type MessageHandler func(msg *Message) error
type wpcSubscriber struct {
	topic          string
	handler        MessageHandler
	workerNum      int
	pulsarConsumer pulsar.Consumer
	messageChannel chan pulsar.ConsumerMessage
	workerPool     *workerpool.Pool[*wpcSubscriber]
	waitGroup      sync.WaitGroup
}

func NewWpcSubscriber(h MessageHandler) *wpcSubscriber {
	return &wpcSubscriber{
		handler: h,
	}
}

func (w *wpcSubscriber) SetWorkerNum(num int) *wpcSubscriber {
	w.workerNum = num
	return w
}

func (w *wpcSubscriber) Subscribe(topic string, consumerOptions ...*pulsar.ConsumerOptions) *wpcSubscriber {
	var c pulsar.Consumer
	var err error
	w.topic = topic
	w.messageChannel = make(chan pulsar.ConsumerMessage)
	if len(consumerOptions) == 0 {
		c, err = instance.GetClient().Subscribe(pulsar.ConsumerOptions{
			Topic:             topic,
			SubscriptionName:  topic + "_subscription",
			Type:              pulsar.Shared,
			ReceiverQueueSize: 1000,
			MessageChannel:    w.messageChannel,
		})
	} else {
		consumerOptions[0].Topic = topic
		consumerOptions[0].SubscriptionName = topic + "_subscription"
		consumerOptions[0].MessageChannel = w.messageChannel
		c, err = instance.GetClient().Subscribe(*consumerOptions[0])
	}
	if err != nil {
		slog.Fatal(err)
	}
	w.pulsarConsumer = c
	return w
}

func (w *wpcSubscriber) Start() {
	if w.workerNum == 0 {
		w.workerNum = 1
	}
	if w.pulsarConsumer == nil {
		slog.Fatal("please subscribe first")
	}
	if w.handler == nil {
		slog.Fatal("please set message handler first")
	}
	w.waitGroup = sync.WaitGroup{}
	w.workerPool = workerpool.New[*wpcSubscriber](workerpool.Config{
		MaxWorkersCount:       uint(w.workerNum),
		MaxIdleWorkerDuration: 5 * time.Second,
	}, workerMessageProcessor)
	instance.subscriberMap.Store(w.topic, w)
	w.waitGroup.Add(1)
	w.workerPool.Exec(w)
	w.waitGroup.Wait()
}

func workerMessageProcessor(w *wpcSubscriber) {
	for cm := range w.messageChannel {
		m := AcquireMessage()
		m.Topic = cm.Message.Topic()
		m.Payload = cm.Message.Payload()
		m.Properties = cm.Message.Properties()
		m.MessageId = cm.Message.ID().Serialize()
		m.PublishTime = cm.Message.PublishTime()
		m.EventTime = cm.Message.EventTime()
		m.RedeliveryCount = int(cm.Message.RedeliveryCount())
		m.ProducerName = cm.Message.ProducerName()
		if err := w.handler(m); err != nil {
			slog.Error(err)
			cm.Nack(cm.Message)
		}
		err := cm.Ack(cm.Message)
		if err != nil {
			slog.Error(err)
		}
		ReleaseMessage(m)
	}
	w.waitGroup.Done()
}

func md5Encode(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func (w *wpcSubscriber) Close() {
	w.pulsarConsumer.Close()
	w.workerPool.Stop()
}
