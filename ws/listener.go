package ws

import (
	"crypto/tls"
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net"
	"net/http"
)

type IWebSocketListener interface {
	Start()
	Stop()
	Addr() string
	IsReady() bool
	SetCerts(certFile, keyfile string)
	SetUpgrade(upgrade interface{})
	Port() int
}

type onConnOpenF func(conn *websocket.Conn, clientIP string)
type onConnCloseF func(conn *websocket.Conn)

type WebSocketListener struct {
	addr        string
	port        int
	certFile    string
	keyFile     string
	wsUpgrade   websocket.Upgrader
	wsListener  net.Listener
	httpServ    *http.Server
	onConnOpen  onConnOpenF
	onConnClose onConnCloseF
}

func NewWebSocketListener(addr string, port int, onOpen onConnOpenF, onClose onConnCloseF) IWebSocketListener {
	return &WebSocketListener{
		addr: addr,
		port: port,
		wsUpgrade: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		onConnOpen:  onOpen,
		onConnClose: onClose,
	}
}

func (wa *WebSocketListener) Start() {
	var (
		err      error
		hostPort string
		path     string
	)
	hostPort = fmt.Sprintf("%s:%d", wa.addr, wa.port)
	path = "/ws"

	wa.wsListener, err = net.Listen("tcp", hostPort)
	if err != nil {
		fmt.Println("ws.listen failed", zap.Error(err))
	}
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("receive connection: ")
		conn, e := wa.wsUpgrade.Upgrade(w, r, nil)
		if e != nil {
			return
		}
		wa.onConnOpen(conn, libs.ClientIP(r))
	})

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	wa.httpServ = &http.Server{
		Addr:      hostPort,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	go func() {
		fmt.Println("gorilla websocket listen start", hostPort, wa.certFile, wa.keyFile)

		if wa.certFile != "" && wa.keyFile != "" {
			err = wa.httpServ.ServeTLS(wa.wsListener, wa.certFile, wa.keyFile)
		} else {
			err = wa.httpServ.Serve(wa.wsListener)
		}
		if err != nil {
			fmt.Println("gorilla websocket listen failed", zap.String("Addr", hostPort), zap.Error(err))
		}
	}()
}

func (wa *WebSocketListener) Stop() {
	err := wa.httpServ.Close()
	if err != nil {
		fmt.Println("websocket close failed", zap.Error(err))
	}
}

func (wa *WebSocketListener) Addr() string {
	return wa.addr
}

func (wa *WebSocketListener) IsReady() bool {
	return wa.Port() != 0
}

func (wa *WebSocketListener) Port() int {
	return wa.port
}

func (wa *WebSocketListener) SetCerts(certFile string, keyFile string) {
	wa.certFile = certFile
	wa.keyFile = keyFile
}

func (wa *WebSocketListener) SetUpgrade(upgrade interface{}) {
	wa.wsUpgrade = upgrade.(websocket.Upgrader)
}
