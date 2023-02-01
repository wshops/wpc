package wpc

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gookit/slog"
	"time"
	"wpc/workerpool"
)

type MessageHandler func(qmsg *pulsar.Message) error
type wpcSubscriber struct {
	topic          string
	handler        MessageHandler
	workerNum      int
	pulsarConsumer pulsar.Consumer
	messageChannel chan pulsar.ConsumerMessage
	workerPool     *workerpool.Pool[*wpcSubscriber]
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
	w.workerPool = workerpool.New[*wpcSubscriber](workerpool.Config{
		MaxWorkersCount:       uint(w.workerNum),
		MaxIdleWorkerDuration: 5 * time.Second,
	}, workerMessageProcessor)
	instance.subscriberMap.Store(w.topic, w)
	w.workerPool.Exec(w)
}

func workerMessageProcessor(w *wpcSubscriber) {
	for cm := range w.messageChannel {
		if err := w.handler(&cm.Message); err != nil {
			slog.Error(err)
			cm.Nack(cm.Message)
		}
		err := cm.Ack(cm.Message)
		if err != nil {
			slog.Error(err)
		}
	}
}

func md5Encode(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func (w *wpcSubscriber) Close() {
	w.pulsarConsumer.Close()
	w.workerPool.Stop()
}
