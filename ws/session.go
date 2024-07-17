package ws

import (
	"errors"
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"sync"
)

type IWebSocketSession interface {
	ID() int64
	SetID(id int64)
	Ok() bool
	GetConn() interface{} // raw conn
	RemoteIP() string
	Send(msg interface{})
	sendLoop()
	Start()
	Close()
	revLoop()
	protectedReadMessage() (interface{}, error)
}
type SessionInit struct{}
type SessionAccepted struct{}
type SessionConnected struct{}
type SessionConnectError struct{}
type SessionClosed struct {
	C int
}
type onSessFunc func(sess IWebSocketSession)
type DecodeFunc func(p []byte) (interface{}, error)
type EncodeFunc func(interface{}) ([]byte, error)
type SendCallback func(interface{})

type Event interface {
	Session() IWebSocketSession
	Message() interface{}
}

type MsgEvent struct {
	Sess IWebSocketSession
	Msg  interface{}
}

func (evt *MsgEvent) Session() IWebSocketSession {
	return evt.Sess
}

func (evt *MsgEvent) Message() interface{} {
	return evt.Msg
}

type OnSessionMessageFunc func(*MsgEvent)

type WebSocketSession struct {
	id               int64
	conn             *websocket.Conn
	remoteIP         string
	exitSync         sync.WaitGroup
	sendQueue        *libs.Pipe
	ok               bool
	capturePanic     bool
	onSessionOpen    onSessFunc
	onSessionClose   onSessFunc
	onSessionMessage OnSessionMessageFunc
	encoder          EncodeFunc
	decoder          DecodeFunc
	sendCallback     SendCallback
}

func GetRemoteAddress(ses *WebSocketSession) (string, bool) {

	if c, ok := ses.GetConn().(*websocket.Conn); ok {
		return c.RemoteAddr().String(), true
	}

	return "", false
}

func newSession(conn *websocket.Conn, clientIP string, onOpen onSessFunc, onClose onSessFunc) *WebSocketSession {
	sess := &WebSocketSession{
		conn:           conn,
		remoteIP:       clientIP,
		sendQueue:      libs.NewPipe(),
		onSessionOpen:  onOpen,
		onSessionClose: onClose,
		capturePanic:   true,
	}
	if sess.remoteIP == "" {
		sess.remoteIP, _ = GetRemoteAddress(sess)
	}

	return sess
}

func (sess *WebSocketSession) ID() int64 {
	return sess.id
}

func (sess *WebSocketSession) SetID(id int64) {
	sess.id = id
}

func (sess *WebSocketSession) EnableCaptureIOPanic(v bool) {
	sess.capturePanic = v
}

func (sess *WebSocketSession) Ok() bool {
	return sess.ok
}

func (sess *WebSocketSession) GetConn() interface{} {
	if sess.conn == nil {
		return nil
	}

	return sess.conn
}

func (sess *WebSocketSession) RemoteIP() string {
	return sess.remoteIP
}

func (sess *WebSocketSession) Start() {
	sess.onSessionOpen(sess)

	sess.exitSync.Add(2)

	go func() {
		sess.exitSync.Wait()
		sess.onSessionClose(sess)
	}()

	go sess.revLoop()
	go sess.sendLoop()

	sess.OnSessionMessage(SessionAccepted{})
	sess.ok = true
}

func (sess *WebSocketSession) Close() {
	sess.ok = false
	sess.sendQueue.Add(nil)
}

func (sess *WebSocketSession) Send(msg interface{}) {
	if !sess.ok {
		return
	}
	sess.sendQueue.Add(msg)
}

func (sess *WebSocketSession) sendLoop() {
	var writeList []interface{}
	for {
		writeList = writeList[0:0]
		exit := sess.sendQueue.Pick(&writeList)

		for _, msg := range writeList {
			sess.SendMessage(msg)
		}

		if exit {
			break
		}
	}

	if sess.conn != nil {
		err := sess.conn.Close()
		if err != nil {
			fmt.Println("session close conn failed: ", err)
		}
		sess.conn = nil
	}

	sess.exitSync.Done()
}

func (sess *WebSocketSession) protectedReadMessage() (msg interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("[WS] IO read panic", zap.Any("Error", e))
			sess.Close()
		}
	}()

	msg, err = sess.ReadMessage()

	return
}

func (sess *WebSocketSession) revLoop() {
	for sess.conn != nil {
		var msg interface{}
		var err error

		if sess.capturePanic {
			msg, err = sess.protectedReadMessage()
		} else {
			msg, err = sess.ReadMessage()
		}

		if err != nil {
			sess.ok = false
			if !libs.IsEOFOrNetReadError(err) {
				//logger.Error("[WS] session closed", zap.Error(err))
			}
			sess.OnSessionMessage(SessionClosed{C: -1})
			break
		}

		sess.OnSessionMessage(msg)
	}

	sess.Close()
	sess.exitSync.Done()
}

func (sess *WebSocketSession) SetEncoder(fn DecodeFunc) {
	sess.decoder = fn
}

func (sess *WebSocketSession) SetDecoder(fn EncodeFunc) {
	sess.encoder = fn
}

func (sess *WebSocketSession) SetSessionEvent(fn OnSessionMessageFunc) {
	sess.onSessionMessage = fn
}

func (sess *WebSocketSession) SetSendCallback(fn SendCallback) {
	sess.sendCallback = fn
}

func (sess *WebSocketSession) ReadMessage() (interface{}, error) {
	var messageType int
	var raw []byte
	var err error
	messageType, raw, err = sess.conn.ReadMessage()
	dataLen := uint32(len(raw))

	if err != nil {
		return nil, err
	}

	if dataLen < 2 {
		return nil, errors.New("packet short size")
	}

	switch messageType {
	case websocket.BinaryMessage:
		if sess.decoder != nil {
			return sess.decoder(raw), nil
		}
	}

	return nil, errors.New("processor: decoder nil")
}

func (sess *WebSocketSession) SendMessage(msg interface{}) {
	if sess.encoder != nil && msg != nil {
		data, err := sess.encoder(msg)
		if err != nil {
			fmt.Println("websocket encode message failed", zap.Error(err))
		}
		err = sess.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			fmt.Println("websocket send message failed", zap.Error(err))
		}
		if sess.sendCallback != nil {
			sess.sendCallback(msg)
		}
	}
}

func (sess *WebSocketSession) OnSessionMessage(msg interface{}) {
	if sess.onSessionMessage != nil && msg != nil {
		sess.onSessionMessage(&MsgEvent{Sess: sess, Msg: msg})
	}
}
