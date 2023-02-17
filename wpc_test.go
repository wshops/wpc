package wpc

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gookit/slog"
	"github.com/wshops/zlog"
	"testing"
	"time"
)

const pulsarUrl = "pulsar+ssl://pulsar.cloud.wshop.info:32247"
const pulsarToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhZG1pbiJ9.n9kWq0OPzkSkD2K29XCdbA9weQaXWkabBk7iLchgb7IAgQt_UmpmcpUTWdIoyPR0h2fgVUMk84DjCvFc_o1zkQMA_SCjE0KW-CkpTfxq1wRRGj3R25env5qL8vbSJOkQtMxY5S6AQ-hYpJUqKIpBZYH01AxFxjg-uWNB65WVbJ7GZFMM7zpMsIKNhMIKkjeSDQlpXHcBZfNuXl5QJnI3-a7QHEEUBn03teUNmXLRxAL6kEPFoSh5dmlyOHiLQCChiRwcv4aqbmCf8Y8oI6K5dGIcGw68xsdjtUu-NbLSMPTc2fKdysfaJJ1vHbKlKC-sY3WtC1O1IWsqswenCeOetQ"

func TestCreateClient(t *testing.T) {
	zlog.New(zlog.LevelDev)
	wt := New(pulsarUrl, zlog.Log(), &pulsar.ClientOptions{
		TLSAllowInsecureConnection: true,
		TLSValidateHostname:        false,
		Authentication:             pulsar.NewAuthenticationToken(pulsarToken),
	})
	t.Cleanup(func() {
		Close()
	})
	if wt == nil {
		t.Fatal("client create failed")
	}
}

func setUpConn() {
	zlog.New(zlog.LevelDev)
	New(pulsarUrl, zlog.Log(), &pulsar.ClientOptions{
		TLSAllowInsecureConnection: true,
		TLSValidateHostname:        false,
		Authentication:             pulsar.NewAuthenticationToken(pulsarToken),
	})
}

func TestCreateProducer(t *testing.T) {
	topic := "test-produce-topic"
	setUpConn()
	t.Cleanup(func() {
		Close()
	})
	Pd(topic)

	v, ok := instance.producerMap.Load(topic)
	if !ok || v == nil {
		t.Error("producer create failed")
	}
}

func TestSendMessage(t *testing.T) {
	setUpConn()
	t.Cleanup(func() {
		Close()
	})
	p := Pd("test-send-topic")
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
		zlog.Log().Info("receive message: ", string(msg.Payload))
		return nil
	}).Subscribe("test-send-topic").
		SetWorkerNum(1).
		Start()
	time.Sleep(time.Second * 5)
	p := Pd("test-send-topic")
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
