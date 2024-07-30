package main

import (
	"context"
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/bwgame666/common/ws"
)

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
	_ = libs.RunForever()
}
