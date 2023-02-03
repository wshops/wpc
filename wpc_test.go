package wpc

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gookit/slog"
	"testing"
	"time"
)

const pulsarUrl = ""
const pulsarToken = ""

func TestCreateClient(t *testing.T) {
	wt := New(pulsarUrl)
	t.Cleanup(func() {
		Close()
	})
	if wt == nil {
		t.Fatal("client create failed")
	}
}

func setUpConn() {
	New(pulsarUrl, &pulsar.ClientOptions{
		Authentication: pulsar.NewAuthenticationToken(pulsarToken),
	})
}

func TestCreateProducer(t *testing.T) {
	setUpConn()
	t.Cleanup(func() {
		Close()
	})
	Pd("persistent://dev-test/test/mq-topic-1")
	v, ok := instance.producerMap.Load("persistent://dev-test/test/mq-topic-1")
	if !ok || v == nil {
		t.Error("producer create failed")
	}
}

func TestSendMessage(t *testing.T) {
	setUpConn()
	t.Cleanup(func() {
		Close()
	})
	p := Pd("persistent://dev-test/test/mq-topic-1")
	mid, err := p.Send(context.Background(), &pulsar.ProducerMessage{
		Payload:   []byte("hello world"),
		EventTime: time.Now(),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log("message sent, id:", mid)
}

func TestConsumeMessage(t *testing.T) {
	setUpConn()
	t.Cleanup(func() {
		Close()
	})
	go NewWpcSubscriber(func(msg *Message) error {
		slog.Info("receive message: ", string(msg.Payload))
		return nil
	}).Subscribe("persistent://dev-test/test/mq-topic-1").
		SetWorkerNum(1).
		Start()
	time.Sleep(time.Second * 5)
	p := Pd("persistent://dev-test/test/mq-topic-1")
	for i := 0; i < 10; i++ {
		_, err := p.Send(context.Background(), &pulsar.ProducerMessage{
			Payload:   []byte("hello world " + fmt.Sprint(i)),
			EventTime: time.Now(),
		})
		if err != nil {
			t.Error(err)
		}
		slog.Info("message sent")
	}
}
