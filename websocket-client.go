package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type SenderWebSocket struct {
	conn *websocket.Conn
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

func (s *SenderWebSocket) sendCode(code []byte) error {
	// #TODO: line endings for data? UTF-8?
	msg := Msg{Type: "code", Data: code}
	return s.conn.WriteJSON(&msg)
}

func (s *SenderWebSocket) sendIdentity(identity *Identity) error {
	msg := Msg{Type: "identity", Data: []byte(identity.DisplayName)}
	return s.conn.WriteJSON(&msg)
}

func (s *SenderWebSocket) Close() {
	s.conn.Close()
}
