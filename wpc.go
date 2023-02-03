package wpc

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gookit/slog"
	"github.com/wshops/wpc/wpclogger"
	"sync"
	"time"
)

type wpc struct {
	pulsarClient  pulsar.Client
	producerMap   sync.Map
	subscriberMap sync.Map
}

var instance *wpc

func New(connectionUrl string, pulsarOptions ...*pulsar.ClientOptions) *wpc {
	var c pulsar.Client
	var err error
	if len(pulsarOptions) == 0 {
		c, err = pulsar.NewClient(pulsar.ClientOptions{
			URL:               connectionUrl,
			OperationTimeout:  30 * time.Second,
			ConnectionTimeout: 30 * time.Second,
			Logger:            wpclogger.NewWpcLogger(),
		})
	} else {
		pulsarOptions[0].URL = connectionUrl
		pulsarOptions[0].Logger = wpclogger.NewWpcLogger()
		c, err = pulsar.NewClient(*pulsarOptions[0])
	}
	if err != nil {
		slog.Fatal(err)
	}
	instance = &wpc{
		pulsarClient: c,
	}
	return instance
}

func Close() {
	if instance != nil {
		instance.producerMap.Range(func(key, value interface{}) bool {
			value.(pulsar.Producer).Close()
			return true
		})
		instance.subscriberMap.Range(func(key, value interface{}) bool {
			value.(*wpcSubscriber).Close()
			return true
		})
		instance.pulsarClient.Close()
	}
}

func (w *wpc) GetProducer(topic string) pulsar.Producer {
	if producer, ok := w.producerMap.Load(topic); ok {
		return producer.(pulsar.Producer)
	}
	p, err := w.pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		slog.Error(err)
		return nil
	}
	w.producerMap.Store(topic, p)
	return p
}

func Pd(topic string) pulsar.Producer {
	return instance.GetProducer(topic)
}

func (w *wpc) GetClient() pulsar.Client {
	return w.pulsarClient
}
