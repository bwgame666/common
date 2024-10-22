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
	ServID   int64  `yaml:"serv_id"`
	Addr     string `yaml:"addr"`
	Port     int    `yaml:"port"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type WebSocket struct {
	listener IWebSocketListener
	sessMgr  ISessionManager

	defaultEncoder EncodeFunc
	defaultDecoder DecodeFunc
	defaultReceive OnSessionMessageFunc

	started bool
	closed  bool
	mu      sync.Mutex
}

func NewWebSocket(conf *WebSocketConfig, initFunc ...ServerOption) *WebSocket {
	webSocketObj := &WebSocket{
		sessMgr:        NewSessionManager(libs.UnixSecs()),
		defaultDecoder: JsonDecoder,
		defaultEncoder: JsonEncoder,
		defaultReceive: nil,
	}
	webSocketObj.listener = NewWebSocketListener(conf.Addr, conf.Port, webSocketObj.OnConnOpen, webSocketObj.OnConnClose)
	if conf.CertFile != "" && conf.KeyFile != "" {
		webSocketObj.listener.SetCerts(conf.CertFile, conf.KeyFile)
	}
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

func (ws *WebSocket) SetDefaultDecoder(fn DecodeFunc) {
	ws.defaultDecoder = fn
}

func (ws *WebSocket) SetDefaultEncoder(fn EncodeFunc) {
	ws.defaultEncoder = fn
}

func (ws *WebSocket) SetSessionEvent(fn OnSessionMessageFunc) {
	ws.defaultReceive = fn
}

func (ws *WebSocket) Broadcast(msg interface{}) {
	ws.sessMgr.ForEach(func(sess IWebSocketSession) bool {
		sess.Send(msg)
		return true
	})
}

func (ws *WebSocket) OnConnOpen(conn *websocket.Conn, clientIP string) {
	fmt.Println("websocket OnConnOpen")
	ses := newSession(conn, clientIP, ws.OnSessionOpen, ws.OnSessionClose)
	ses.SetDecoder(ws.defaultDecoder)
	ses.SetEncoder(ws.defaultEncoder)
	if ws.defaultReceive != nil {
		ses.SetSessionEvent(ws.defaultReceive)
	}
	ses.Start()
}

func (ws *WebSocket) OnConnClose(conn *websocket.Conn) {
	fmt.Println("websocket OnConnClose")
}

func (ws *WebSocket) OnSessionOpen(sess IWebSocketSession) {
	fmt.Println("websocket OnSessionOpen")
	ws.AddSession(sess)
}

func (ws *WebSocket) OnSessionClose(sess IWebSocketSession) {
	fmt.Println("websocket OnSessionClose")
	ws.RemoveSession(sess)
}
