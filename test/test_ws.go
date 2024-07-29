package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwgame666/common/ws"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

func RunForever() error {
	var err error
	signals := []os.Signal{os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT,
		syscall.SIGUSR2, syscall.SIGKILL}
	sCtx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(sCtx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case sig := <-signalChan:
			fmt.Println("receive sig: ", sig)
			switch sig {
			case syscall.SIGUSR2:
				// TODO
				return nil
			default:
				return nil
			}
		}
	})
	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		cancel()
		return err
	}
	cancel()
	return nil
}

type TestMsg struct {
	Hello string `json:"hello"`
}

func OnReceive(msg *ws.MsgEvent) {
	switch msg.Message().(type) {
	case []byte:
		fmt.Println("new msg: ", msg.Session().ID(), string(msg.Message().([]byte)))
		response := &TestMsg{
			Hello: "hello world!",
		}
		msg.Session().Send(response)
	case ws.SessionAccepted:
		fmt.Println("SessionAccepted: ", msg.Session().ID())
	case ws.SessionClosed:
		fmt.Println("SessionClosed: ", msg.Session().ID())
	default:
		fmt.Println("unknown type msg: ", msg.Session().ID())
	}

}

func main() {
	wsConf := &ws.WebSocketConfig{
		ServID: 10086,
		Addr:   "127.0.0.1",
		Port:   10086,
	}
	webS := ws.NewWebSocket(wsConf)
	webS.SetSessionEvent(OnReceive)
	err := webS.Start(context.Background())
	if err != nil {
		fmt.Println("start websocket failed")
		return
	}
	_ = RunForever()
}
