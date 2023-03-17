package wpc

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/wshops/wpc/wpcl"
	"github.com/wshops/wpc/wpcm"
	"go.uber.org/zap"
	"time"
)

type Wpc struct {
	log         *zap.SugaredLogger
	tenantId    string
	producerMap map[string]wpcm.Producer
	consumerMap map[string]wpcm.Consumer
}

var instance *Wpc

func New(connectionUrl string, tenant string, logger *zap.SugaredLogger, pulsarOptions ...*pulsar.ClientOptions) *Wpc {
	if instance != nil {
		instance.log.Warn("wpc instance already exists")
		return instance
	}
	var c *pulsar.ClientOptions
	if len(pulsarOptions) == 0 {
		c = &pulsar.ClientOptions{
			URL:               connectionUrl,
			OperationTimeout:  30 * time.Second,
			ConnectionTimeout: 30 * time.Second,
			Logger:            wpcl.NewBlackHoleLogger(logger),
		}
	} else {
		c = pulsarOptions[0]
		c.URL = connectionUrl
		if c.Logger == nil {
			c.Logger = wpcl.NewBlackHoleLogger(logger)
		}
	}
	err := wpcm.InitMessaging(1, c)
	if err != nil {
		logger.Fatal(err)
	}
	instance = &Wpc{
		log:         logger,
		tenantId:    tenant,
		producerMap: make(map[string]wpcm.Producer),
		consumerMap: make(map[string]wpcm.Consumer),
	}
	return instance
}

func RegisterProducer(namespace string, topic string) {
	p, err := wpcm.CreateProducer(instance.tenantId, namespace, topic)
	if err != nil {
		instance.log.Error(err)
		return
	}
	instance.producerMap[BuildTopicPath(instance.tenantId, namespace, topic)] = p
}

func GetProducer(namespace string, topic string) wpcm.Producer {
	p := instance.producerMap[BuildTopicPath(instance.tenantId, namespace, topic)]
	if p == nil {
		instance.log.Error("Producer not found")
		return nil
	}
	return p
}

func RegisterConsumer(namespace string, topic string, handler wpcm.Handler) {
	consumer, err := wpcm.CreateSingleTopicConsumer(instance.tenantId, namespace, topic, handler, wpcm.ConsumerOpts{
		SubscriptionName: topic + "_subscription",
		RetryEnabled:     true,
		InitialPosition:  wpcm.Latest,
	})
	if err != nil {
		instance.log.Error(err)
		return
	}
	instance.consumerMap[BuildTopicPath(instance.tenantId, namespace, topic)] = consumer
}

func GetConsumer(namespace string, topic string) wpcm.Consumer {
	c := instance.consumerMap[BuildTopicPath(instance.tenantId, namespace, topic)]
	if c == nil {
		instance.log.Error("Consumer not found")
		return nil
	}
	return c
}

func Close() {
	for _, p := range instance.producerMap {
		p.Stop()
	}
	for _, c := range instance.consumerMap {
		err := c.Stop()
		if err != nil {
			instance.log.Error(err)
		}
	}
	wpcm.Cleanup()
}

func BuildTopicPath(tenantName, namespace, topicName string) string {
	return fmt.Sprintf("persistent://%s/%s/%s", tenantName, namespace, topicName)
}
