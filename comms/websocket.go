package comms

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	WS_UPGRADER = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type FnWsConn func(*websocket.Conn) error

func WsUpgrade(w http.ResponseWriter, r *http.Request, fn FnWsConn) error {
	conn, err := WS_UPGRADER.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("client couldn't upgrade: %w", err)
	}
	defer conn.Close()

	err = fn(conn)
	if err != nil {
		return fmt.Errorf("client connection raised an error: %w", err)
	}

	return nil
}
