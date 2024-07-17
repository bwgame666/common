package ws

import (
	"context"
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/gorilla/websocket"
	"sync"
)

type ServerOption func(webSocketObj *WebSocket)

type WebSocketConfig struct {
	ServID int64  `yaml:"serv_id"`
	Addr   string `yaml:"addr"`
	Port   int    `yaml:"port"`
}

type WebSocket struct {
	listener IWebSocketListener
	sessMgr  ISessionManager

	started bool
	closed  bool
	mu      sync.Mutex
}

func NewWebSocket(conf *WebSocketConfig, initFunc ...ServerOption) *WebSocket {
	webSocketObj := &WebSocket{
		sessMgr: NewSessionManager(libs.UnixSecs()),
	}
	webSocketObj.listener = NewWebSocketListener(conf.Addr, conf.Port, webSocketObj.OnConnOpen, webSocketObj.OnConnClose)
	for _, iFunc := range initFunc {
		iFunc(webSocketObj)
	}
	return webSocketObj
}

func (ws *WebSocket) Start(ctx context.Context) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.started {
		return fmt.Errorf("websocket already started")
	}
	ws.started = true

	ws.listener.Start()

	return nil
}

func (ws *WebSocket) Stop(ctx context.Context) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return nil
	}
	ws.closed = true
	ws.started = false
	ws.listener.Stop()
	ws.sessMgr.CloseAll()

	return nil
}

func (ws *WebSocket) AddSession(sess IWebSocketSession) {
	ws.sessMgr.Add(sess)
}

func (ws *WebSocket) GetSession(sessId int64) IWebSocketSession {
	return ws.sessMgr.Get(sessId)
}

func (ws *WebSocket) RemoveSession(sess IWebSocketSession) {
	//sess.ProcData(sess, SessionClosed{})
	ws.sessMgr.Remove(sess)
}

func (ws *WebSocket) CloseSession(sessId int64) {
	sess := ws.sessMgr.Get(sessId)
	if sess != nil {
		sess.Close()
	}
}

func (ws *WebSocket) Broadcast(msg interface{}) {
	ws.sessMgr.ForEach(func(sess IWebSocketSession) bool {
		sess.Send(msg)
		return true
	})
}

func (ws *WebSocket) OnConnOpen(conn *websocket.Conn, clientIP string) {
	ses := newSession(conn, clientIP, ws.OnSessionOpen, ws.OnSessionClose)
	ses.Start()
}

func (ws *WebSocket) OnConnClose(conn *websocket.Conn) {

}

func (ws *WebSocket) OnSessionOpen(sess IWebSocketSession) {
	ws.AddSession(sess)
}

func (ws *WebSocket) OnSessionClose(sess IWebSocketSession) {
	ws.RemoveSession(sess)
}
