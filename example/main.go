package main

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/wshops/wpc/v2/wpcl"
	"github.com/wshops/wpc/v2/wpcm"
	"github.com/wshops/zlog"
	"os"
	"os/signal"
	"syscall"
)

const pulsarUrl = ""
const pulsarToken = ""

func main() {
	zlog.New(zlog.LevelDev)
	wpc.New(pulsarUrl, "wpc_test", zlog.Log(), &pulsar.ClientOptions{
		TLSAllowInsecureConnection: true,
		TLSValidateHostname:        false,
		Authentication:             pulsar.NewAuthenticationToken(pulsarToken),
		Logger:                     wpcl.NewBlackHoleLogger(zlog.Log()),
	})
	defer wpc.Close()
	wpc.RegisterProducer("render", "testmsg")
	wpc.RegisterConsumer("render", "testmsg", &TestRenderHandler{})
	err := wpc.GetConsumer("render", "testmsg").Start()
	if err != nil {
		zlog.Log().Error(err)
		return
	}

	go func() {
		counter := 0
		for true {
			err := wpc.GetProducer("render", "testmsg").PublishOne(&wpcm.Message{
				Payload: []byte(fmt.Sprintf("test message GG%d", counter)),
			})
			zlog.Log().Infof("[SENT MSG] GG%d", counter)
			if err != nil {
				return
			}
			counter++
		}
	}()

	NewShutdownHook().Close(func() {
		wpc.Close()
		err := zlog.Log().Sync()
		if err != nil {
			return
		}
	})
}

type TestRenderHandler struct {
}

func (t *TestRenderHandler) HandleMessage(message *wpcm.Message) *wpcm.RetryMessage {
	zlog.Log().Infof("[NEW MSG] Message: %s", string(message.Payload))
	return nil
}

var _ Hook = (*hook)(nil)

// Hook a graceful shutdown hook, default with signals of SIGINT and SIGTERM
type Hook interface {
	// WithSignals add more signals into hook
	WithSignals(signals ...syscall.Signal) Hook

	// Close register shutdown handles
	Close(funcs ...func())
}

type hook struct {
	ctx chan os.Signal
}

// NewHook create a Hook instance
func NewShutdownHook() Hook {
	hook := &hook{
		ctx: make(chan os.Signal, 1),
	}

	return hook.WithSignals(syscall.SIGINT, syscall.SIGTERM)
}

func (h *hook) WithSignals(signals ...syscall.Signal) Hook {
	for _, s := range signals {
		signal.Notify(h.ctx, s)
	}

	return h
}

func (h *hook) Close(funcs ...func()) {
	select {
	case <-h.ctx:
	}
	signal.Stop(h.ctx)

	for _, f := range funcs {
		f()
	}
}
