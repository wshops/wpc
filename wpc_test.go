package wpc

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gookit/slog"
	"go.uber.org/zap"
	"testing"
	"time"
)

const pulsarUrl = "pulsar+ssl://wshop-test-1421d17d-7ede-40e5-994e-01f15c705276.gcp-shared-gcp-usce1-martin.streamnative.g.snio.cloud:6651"
const pulsarToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ik5rRXdSVVU1TUVOQlJrWTJNalEzTVRZek9FVkZRVVUyT0RNME5qUkRRVEU1T1VNMU16STVPUSJ9.eyJodHRwczovL3N0cmVhbW5hdGl2ZS5pby91c2VybmFtZSI6ImFueHVhbnppQHctc2hvcHMuY29tIiwiaXNzIjoiaHR0cHM6Ly9hdXRoLnN0cmVhbW5hdGl2ZS5jbG91ZC8iLCJzdWIiOiJhdXRoMHw2M2RjNTNjYWFlNDYxNzk3NTQ2MWQ1NDciLCJhdWQiOlsidXJuOnNuOnB1bHNhcjpvLXlmMW11OndzaG9wLXRlc3QiLCJodHRwczovL3N0cmVhbW5hdGl2ZS1wcm9kLXVzLmF1dGgwLmNvbS91c2VyaW5mbyJdLCJpYXQiOjE2NzUzODU0OTcsImV4cCI6MTY3NTk5MDI5NywiYXpwIjoiQUpZRWRIV2k5RUZla0VhVVhrUFdBMk1xUTNscTFOckkiLCJzY29wZSI6Im9wZW5pZCBlbWFpbCBhZG1pbiIsInBlcm1pc3Npb25zIjpbImFkbWluIl19.GOdFu8HD9BHEeIUa8DNl-11YE_phvnOrE5IMvgvuzOcBPVZ96a3vvbaYoM3tFdmoEqCoYmzf-rvt5Lb0h_ypCwJwdnauUqO-3oFY9V_imE3g9BtNtysK6LOQeEjo9y9LWvurNb1b7PCP4bUS5dTXLoD20wgHbtEkSOvYhYXJuNwqIO1RPw766DAcrSPKrDKZCH5Ck-FZLgoIEalqYHwzS-4rvd4CndpKfsOGmxbcxouEKKlRFiYr3RFfnZFQwxoyC8PYnXDkeQA26yqx_cxHz1tGauNEUcU03ucSblj40Ipd3OvhrpKEov7HP9e37muAptcSYiHpGqj_ObqTE22How"

func TestCreateClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wt := New(pulsarUrl, logger.Sugar(), &pulsar.ClientOptions{
		Authentication: pulsar.NewAuthenticationToken(pulsarToken),
	})
	t.Cleanup(func() {
		Close()
	})
	if wt == nil {
		t.Fatal("client create failed")
	}
}

func setUpConn() {
	logger, _ := zap.NewDevelopment()
	New(pulsarUrl, logger.Sugar(), &pulsar.ClientOptions{
		Authentication: pulsar.NewAuthenticationToken(pulsarToken),
	})
}

func TestCreateProducer(t *testing.T) {
	topic := "persistent://dev-test/test/mq-topic-1"
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
