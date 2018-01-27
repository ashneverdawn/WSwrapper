package WSwrapper

import (
	"github.com/gorilla/websocket"
	"time"
	"net"
	"net/url"
	"net/http"
)

type WSconn struct{ 
	*websocket.Conn
}
func (ws WSconn) Read(b []byte) (int, error) {
	_, msg, err := ws.ReadMessage()
	if err != nil {
		return 0, err
	}
	if msg != nil {
		for i := 0; i < len(msg) && i < len(b); i++ {
			b[i] = msg[i]
		}
	}
	length := len(msg)
	if len(b) < len(msg) {
		length = len(b)
	}
	return length, nil
}
func (ws WSconn) Write(b []byte) (int, error) {
	err := ws.WriteMessage(websocket.TextMessage, b)
	return len(b), err
}
func (ws WSconn) SetDeadline(t time.Time) (error) {
	return nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

type Dialer struct {}

func (d *Dialer) Dial(network, address string) (net.Conn, error){
	u := url.URL{Scheme: "ws", Host: address, Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return WSconn{conn}, err
}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
}
func Serve(ln net.Listener, address string, handler func(net.Conn) ) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { 
			println(err)
			panic(err.Error())
		}
		handler(WSconn{conn})
	})
	server := &http.Server{Addr: address}
	e := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
	if e != nil {
		panic(e)
	}
}