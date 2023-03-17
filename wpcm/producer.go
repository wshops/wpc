package wpcm

import "github.com/apache/pulsar-client-go/pulsar"

// Producer
// @Description: Interface for pulsar producer.
type Producer interface {
	// Publish
	// @Description: This function publishes the messages to the topic.
	// @Param []*Message
	Publish([]*Message) error

	// PublishOne
	// @Description: This function publishes the message to the topic.
	// @Param *Message
	PublishOne(*Message) error

	// Stop
	// @Description: This will close the producer client.
	Stop()

	// Stats
	// @Description: This function will provide stats of the messages produced.
	Stats() Stats
}

type producer struct {
	pulsarp pulsar.Producer
	stats   Stats
	stopped bool
}
