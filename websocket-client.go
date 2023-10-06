package main

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/creativenucleus/bytejammer/machines"
	"github.com/gorilla/websocket"
)

type SenderWebSocket struct {
	conn    *websocket.Conn
	wsMutex sync.Mutex
}

// Ensure you:
//
//	defer sender.Close()
func NewWebSocketClient(host string, port int) (*SenderWebSocket, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/ws-bytejam",
		//		User:   userInfo,
	}
	log.Printf("-> Connecting to %s", u.String())

	s := SenderWebSocket{}
	var err error
	s.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *SenderWebSocket) Close() {
	s.conn.Close()
}

func (s *SenderWebSocket) sendCode(ts machines.TicState) error {
	// #TODO: line endings for data? UTF-8?
	msg := Msg{Type: "tic-state", TicState: ts}
	return s.sendData(&msg)
}

func (s *SenderWebSocket) sendIdentity(identity *Identity) error {
	msg := Msg{Type: "identity", Identity: identity.DisplayName}
	return s.sendData(&msg)
}

func (s *SenderWebSocket) sendData(data interface{}) error {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	return s.conn.WriteJSON(data)
}
