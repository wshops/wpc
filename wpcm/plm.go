package wpcm

// copied and modified from https://github.com/blueshift-labs/pulsarlib-go/blob/master/pulsarlib/messaging.go
import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// RetryMessage
// @Description: Struct for retrying consumed message due to failure. Client can return this struct to ensure the message will be enqueued for a given RetryAfter duration.
type RetryMessage struct {
	RetryAfter time.Duration
}

// Message
// @Description: Struct for pulsar message.
type Message struct {
	Payload      []byte
	Properties   map[string]string
	DeliverAfter time.Duration
}

// Stats
// @Description: Struct for pulsar message stats.
type Stats struct {
	TotalMessages uint64
}

// IncrementMessageCount
// @Description: This function will increment the total messages consumed.
// @Param messages
func (s Stats) IncrementMessageCount(messages uint64) {
	atomic.AddUint64(&s.TotalMessages, messages)
}

var msging *messaging

// Internal structs
type messageItem struct {
	message pulsar.Message
	wg      *sync.WaitGroup
	handler Handler
	pulsarc pulsar.Consumer
}

type messaging struct {
	nWorkers  int
	messageCh chan *messageItem
	client    pulsar.Client
}

func (p *producer) Publish(msgs []*Message) error {
	if p.stopped {
		return fmt.Errorf("Producer is stopped")
	}
	for _, msg := range msgs {
		pulsarMsg := &pulsar.ProducerMessage{
			Payload:      msg.Payload,
			Properties:   msg.Properties,
			DeliverAfter: msg.DeliverAfter,
		}
		_, err := p.pulsarp.Send(context.Background(), pulsarMsg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *producer) PublishOne(msg *Message) error {
	if p.stopped {
		return fmt.Errorf("Producer is stopped")
	}

	pulsarMsg := &pulsar.ProducerMessage{
		Payload:      msg.Payload,
		Properties:   msg.Properties,
		DeliverAfter: msg.DeliverAfter,
	}
	_, err := p.pulsarp.Send(context.Background(), pulsarMsg)
	if err != nil {
		return err
	}

	return nil
}

func (p *producer) Stats() Stats {
	return p.stats
}

func (p *producer) Stop() {
	if p.stopped {
		return
	}

	p.pulsarp.Close()
	p.stopped = true
}

func (m *messaging) processMessageWorker() {
	for messageItem := range m.messageCh {
		m := &Message{
			Payload:    messageItem.message.Payload(),
			Properties: messageItem.message.Properties(),
		}
		retry := messageItem.handler.HandleMessage(m)
		if retry != nil {
			if retry.RetryAfter == 0 {
				messageItem.pulsarc.Nack(messageItem.message)
			} else {
				messageItem.pulsarc.ReconsumeLater(messageItem.message, retry.RetryAfter)
			}
		} else {
			messageItem.pulsarc.Ack(messageItem.message)
		}
		messageItem.wg.Done()
	}
}

func (c *consumer) commit() {
	//NoOp for pulsar as consumer are in shared mode and send the acknowledgement individually
}

func (c *consumer) pauseWait() {
	<-c.unpauseCh
}

func (c *consumer) messageFetcher() {
	for {
		ctx, canc := context.WithCancel(c.ctx)
		message, err := c.pulsarc.Receive(ctx)
		if err != nil && err != context.Canceled {
			log.Printf("Error occured in fetching a message. Error: %v", err)
		}
		canc()

		messageItem := &messageItem{
			message: message,
			wg:      c.messageWg,
			handler: c.handler,
			pulsarc: c.pulsarc,
		}
		//Message can be nil in case of error
		if message != nil {
			c.messageWg.Add(1)
			c.stats.IncrementMessageCount(1)

			msging.messageCh <- messageItem
		}

		//Check for a pause signal
		if c.pauseConsumer {
			//Let the fetched messages flush
			c.messageWg.Wait()
			c.commit()

			//Acknowledge the pause signal
			c.consumerPausedCh <- true
			c.pauseWait()
		}

		//Check for messageFetcher to be stopped
		if c.stopConsumer {
			//Let the fetched messages flush
			c.messageWg.Wait()
			c.commit()
			c.consumerStopWg.Done()
			return
		}

	}
}

func (c *consumer) Start() error {
	if c.stopConsumer {
		return fmt.Errorf("Cannot start a stopped consumer")
	}

	if c.consumerRunning {
		//Consumer is already running
		return nil
	}

	//Start the message fetcher
	go c.messageFetcher()
	c.consumerRunning = true
	return nil
}

func (c *consumer) Stop() error {
	if c.stopConsumer {
		//Consumer is already stopped
		return nil
	}

	c.consumerStopWg.Add(1)
	c.stopConsumer = true
	c.canc()

	c.consumerStopWg.Wait()
	c.consumerRunning = false

	c.pulsarc.Close()
	return nil
}

func (c *consumer) Unsubscribe() error {
	if c.consumerRunning {
		//Consumer is running. Stop first
		return nil
	}

	return c.pulsarc.Unsubscribe()
}

func (c *consumer) Pause() {
	if c.pauseConsumer {
		//Consumer is already paused. Return
		return
	}
	c.pauseConsumer = true
	//Wait for the pause to be acknowledged
	<-c.consumerPausedCh
}

func (c *consumer) Unpause() {
	if !c.pauseConsumer {
		//Consumer is not paused
		return
	}

	c.pauseConsumer = false
	c.unpauseCh <- true
}

func (c *consumer) Stats() Stats {
	return c.stats
}

// InitMessaging
// @Description: This API will initialize the messaging channel.
//
//	It will do all the connection initialization.
//	workerCount is the number of message processing workers.
//
// @Param workerCount
func InitMessaging(workerCount int, pulsarClientOptions *pulsar.ClientOptions) error {
	msging = &messaging{
		nWorkers:  workerCount,
		messageCh: make(chan *messageItem, workerCount*2),
	}

	//Start the processMessage workers
	for i := 0; i < workerCount; i++ {
		go msging.processMessageWorker()
	}
	(*pulsarClientOptions).OperationTimeout = 30 * time.Second
	(*pulsarClientOptions).ConnectionTimeout = 30 * time.Second
	client, err := pulsar.NewClient(*pulsarClientOptions)
	if err != nil {
		return fmt.Errorf("could not instantiate Pulsar client: %v", err)
	}
	msging.client = client
	return nil
}

func Cleanup() {
	if msging == nil {
		return
	}
	msging.client.Close()
	msging = nil
}

type InitialPosition int

const (
	// Latest position which means the start consuming position will be the last message
	Latest InitialPosition = iota

	// Earliest position which means the start consuming position will be the first message
	Earliest
)

type ConsumerOpts struct {
	SubscriptionName string
	RetryEnabled     bool
	InitialPosition  InitialPosition
}

func toInitialPosition(p InitialPosition) pulsar.SubscriptionInitialPosition {
	switch p {
	case Latest:
		return pulsar.SubscriptionPositionLatest
	case Earliest:
		return pulsar.SubscriptionPositionEarliest
	}

	return pulsar.SubscriptionPositionEarliest
}

// CreateConsumer
// @Description: This API will create a Consumer for a particular topic.
//
//	The handler passed should implement the Handler interface from this module.
//	The consumer will create the subscription and be in a passive state until Start() is called.
//	The consumer can be Paused and Unpaused at any point.
//	The commitInterval used to commit messages after every n messages are consumed.
//	The Pause() function will flushout the already received messages and pause receiving any further messages.
//	The Unpause() function will resume receiving messages.
//	The Stop() function will flush existing messages and stop the consumer. It won't delete the subscription.
//	The Unsubscribe() function can be used if subscription needs to be deleted.
//	The Stats() function provides the stats for messages consumed.
//	Creating multiple instances of Consumer for same topic will deliver message to only one of the instances.
//	Inorder to recreate a Consumer for same topic make sure Stop() is called on old Consumer instance.
//
// @Param tenantID
// @Param namespace
// @Param topics
// @Param handler
// @Param opts
func CreateConsumer(tenantID, namespace string, topics []string, handler Handler, opts ConsumerOpts) (Consumer, error) {
	//Check if InitMessaging was done prior to this call
	if msging == nil {
		return nil, fmt.Errorf("InitMessaging not called yet")
	}

	topicArr := []string{}
	for _, tp := range topics {
		topicArr = append(topicArr, fmt.Sprintf("persistent://%s/%s/%s", tenantID, namespace, tp))
	}
	consumerOptions := pulsar.ConsumerOptions{
		Topics:                      topicArr,
		SubscriptionName:            opts.SubscriptionName,
		Type:                        pulsar.Shared,
		SubscriptionInitialPosition: toInitialPosition(opts.InitialPosition),
	}
	c, err := msging.client.Subscribe(consumerOptions)
	if err != nil {
		return nil, fmt.Errorf("Error in subscribing to the topics. Error %v", err)
	}

	ctx, canc := context.WithCancel(context.Background())

	consumer := &consumer{
		topics:  topics,
		pulsarc: c,
		stats:   Stats{},
		handler: handler,

		ctx:  ctx,
		canc: canc,

		pauseConsumer:    false,
		consumerPausedCh: make(chan bool, 1),
		unpauseCh:        make(chan bool, 1),

		messageWg:      &sync.WaitGroup{},
		consumerStopWg: &sync.WaitGroup{},
	}
	return consumer, nil
}

// CreateSingleTopicConsumer
// @Description: This API will create a Consumer for a particular topic.
//
//	 	The handler passed should implement the Handler interface from this module.
//		The consumer will create the subscription and be in a passive state until Start() is called.
//		The consumer can be Paused and Unpaused at any point.
//		The commitInterval used to commit messages after every n messages are consumed.
//		The Pause() function will flushout the already received messages and pause receiving any further messages.
//		The Unpause() function will resume receiving messages.
//		The Stop() function will flush existing messages and stop the consumer. It won't delete the subscription.
//		The Unsubscribe() function can be used if subscription needs to be deleted.
//		The Stats() function provides the stats for messages consumed.
//		Creating multiple instances of Consumer for same topic will deliver message to only one of the instances.
//		Inorder to recreate a Consumer for same topic make sure Stop() is called on old Consumer instance.
//		retryEnabled will let the consumer retry message in case of HandleMessage return `RetryMessage` struct.
//
// @Param tenantID
// @Param namespace
// @Param topic
// @Param handler
// @Param opts
func CreateSingleTopicConsumer(tenantID, namespace, topic string, handler Handler, opts ConsumerOpts) (Consumer, error) {
	//Check if InitMessaging was done prior to this call
	if msging == nil {
		return nil, fmt.Errorf("InitMessaging not called yet")
	}

	topicURI := fmt.Sprintf("persistent://%s/%s/%s", tenantID, namespace, topic)

	consumerOptions := pulsar.ConsumerOptions{
		Topic:                       topicURI,
		SubscriptionName:            opts.SubscriptionName,
		Type:                        pulsar.Shared,
		SubscriptionInitialPosition: toInitialPosition(opts.InitialPosition),
	}

	if opts.RetryEnabled {
		// We wanted to retry message for 1 minute before it appended at the back of the DLQ topic.
		maxDeliveries := uint32(60) // 1 min
		consumerOptions.RetryEnable = true
		consumerOptions.NackRedeliveryDelay = 1 * time.Second
		consumerOptions.DLQ = &pulsar.DLQPolicy{
			MaxDeliveries:    maxDeliveries,
			DeadLetterTopic:  topicURI,
			RetryLetterTopic: topicURI,
		}
	}

	c, err := msging.client.Subscribe(consumerOptions)
	if err != nil {
		return nil, fmt.Errorf("Error in subscribing to the topics. Error %v", err)
	}

	ctx, canc := context.WithCancel(context.Background())

	consumer := &consumer{
		topics:  []string{topic},
		pulsarc: c,
		stats:   Stats{},
		handler: handler,

		ctx:  ctx,
		canc: canc,

		pauseConsumer:    false,
		consumerPausedCh: make(chan bool, 1),
		unpauseCh:        make(chan bool, 1),

		messageWg:      &sync.WaitGroup{},
		consumerStopWg: &sync.WaitGroup{},
	}
	return consumer, nil
}

// CreateRegexConsumer
// @Description: This API will create a Consumer for a topics matching the topics pattern.
//
//	The handler passed should implement the Handler interface from this module.
//	The consumer will create the subscription and be in a passive state until Start() is called.
//	The consumer can be Paused and Unpaused at any point.
//	The Pause() function will flushout the already received messages and pause receiving any further messages.
//	The Unpause() function will resume receiving messages.
//	The Stop() function will flush existing messages and stop the consumer. It won't delete the subscription.
//	The Unsubscribe() function can be used if subscription needs to be deleted.
//	The Stats() function provides the stats for messages consumed.
//	Creating multiple instances of Consumer for same topic will deliver message to only one of the instances.
//	Inorder to recreate a Consumer for same topic make sure Stop() is called on old Consumer instance.
//
// @Param tenantID
// @Param namespace
// @Param topicsPattern
// @Param handler
// @Param opts
func CreateRegexConsumer(tenantID, namespace, topicsPattern string, handler Handler, opts ConsumerOpts) (Consumer, error) {
	//Check if InitMessaging was done prior to this call
	if msging == nil {
		return nil, fmt.Errorf("InitMessaging not called yet")
	}

	c, err := msging.client.Subscribe(pulsar.ConsumerOptions{
		TopicsPattern:               topicsPattern,
		SubscriptionName:            opts.SubscriptionName,
		Type:                        pulsar.Shared,
		SubscriptionInitialPosition: toInitialPosition(opts.InitialPosition),
	})
	if err != nil {
		return nil, fmt.Errorf("Error in subscribing to the topics. Error %v", err)
	}

	ctx, canc := context.WithCancel(context.Background())

	consumer := &consumer{
		topicsPattern: topicsPattern,
		pulsarc:       c,
		stats:         Stats{},
		handler:       handler,

		ctx:  ctx,
		canc: canc,

		pauseConsumer:    false,
		consumerPausedCh: make(chan bool, 1),
		unpauseCh:        make(chan bool, 1),

		messageWg:      &sync.WaitGroup{},
		consumerStopWg: &sync.WaitGroup{},
	}
	return consumer, nil
}

// CreateProducer
// @Description: This API will create a Producer for a particular topic. The Producer instance can be used to Publish messages to the topic.
// @Param tenantID
// @Param namespace
// @Param topic
func CreateProducer(tenantID string, namespace string, topic string) (Producer, error) {
	//Check if InitMessaging was done prior to this call
	if msging == nil {
		return nil, fmt.Errorf("InitMessaging not called yet")
	}

	topicPath := fmt.Sprintf("persistent://%s/%s/%s", tenantID, namespace, topic)
	p, err := msging.client.CreateProducer(pulsar.ProducerOptions{
		Topic: topicPath,
		// We wanted to send error in case queue is full, this will give the sender a chance to requeue or retry msg
		DisableBlockIfQueueFull: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Error in creating producer. Error %v", err)
	}

	producer := &producer{
		pulsarp: p,
		stats:   Stats{},
	}
	return producer, nil
}
