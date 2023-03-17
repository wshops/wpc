package wpcm

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"sync"
)

// Handler
// @Description: Interface for handling consumed pulsar messages.
type Handler interface {
	// HandleMessage
	// @Description: Handle consumed pulsar message. 'RetryMessage' is used to negatively ack message and requeued to retry after some delay.
	// @Param *Message
	HandleMessage(*Message) *RetryMessage
}

// Consumer
// @Description: Interface for pulsar consumer.
type Consumer interface {
	// Start
	// @Description: This function will start the consumption of messages.
	Start() error

	// Stop
	// @Description: This function will flush any existing messages and stop the consumer client.
	Stop() error

	// Unsubscribe
	// @Description: This function will delete the subscription created by the consumer.
	Unsubscribe() error

	// Pause
	// @Description: This function will flush existing messages and pause further consumption.
	Pause()

	// Unpause
	// @Description: This function will unpause the message consumption.
	Unpause()

	// Stats
	// @Description: This function will provide stats of the messages consumed.
	Stats() Stats
}

type consumer struct {
	topics        []string
	topicsPattern string
	pulsarc       pulsar.Consumer
	stats         Stats
	handler       Handler
	//Context to manage the consumer
	ctx  context.Context
	canc context.CancelFunc

	pauseConsumer bool
	//This channel will be used for acknowledging pause
	consumerPausedCh chan bool
	//Unpause channel will be used to wait for pause to end
	unpauseCh chan bool

	//Flags
	stopConsumer    bool
	consumerRunning bool

	//Waitgroup for tracking messages processed and stop of consumer.
	messageWg      *sync.WaitGroup
	consumerStopWg *sync.WaitGroup
}
