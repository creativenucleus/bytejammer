package main

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// #TODO
/*
const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)
*/

type WebSocketLink struct {
	conn    *websocket.Conn
	wsMutex sync.Mutex
}

// Ensure you:
//
//	defer sender.Close()
func NewWebSocketLink(host string, port int, path string) (*WebSocketLink, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   path,
	}
	log.Printf("-> Connecting to %s", u.String())

	s := WebSocketLink{}
	var err error
	s.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *WebSocketLink) Close() error {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := s.sendControlSignal(websocket.CloseMessage, msg, time.Second)
	if err != nil {
		// #TODO: log - though I don't know if we should still try close?
	}

	return s.conn.Close()
}

func (s *WebSocketLink) sendControlSignal(messageType int, data []byte, byDuration time.Duration) error {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	return s.conn.WriteControl(websocket.CloseMessage, data, time.Now().Add(byDuration))
}

func (s *WebSocketLink) sendData(data interface{}) error {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	return s.conn.WriteJSON(data)
}
