package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer/machines"
)

type ClientWS struct {
	ws *SenderWebSocket
}

type ClientServerStatus struct {
	isConnected bool
}

func startClientServerConn(host string, port int, identity *Identity, chServerStatus chan ClientServerStatus) error {
	chServerStatus <- ClientServerStatus{isConnected: false}
	cws := ClientWS{}
	// Keep running until we make a connection
	for {
		// #TODO: This is not the right construction
		var err error
		cws.ws, err = clientOpenConnection(host, port)
		if err != nil {
			//chServerStatus <- false
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		break
	}
	defer cws.ws.Close()
	chServerStatus <- ClientServerStatus{isConnected: true}

	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return err
	}
	defer m.Shutdown()

	// #TODO: shift import / export to *Machine?
	go cws.clientWsReader(m.Tic)
	go cws.clientWsWriter(m.Tic, identity)

	// Lock #TODO: use a channel to escape
	for {
	}
}

func clientOpenConnection(host string, port int) (*SenderWebSocket, error) {
	ws, err := NewWebSocketClient(host, port)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (cws *ClientWS) clientWsReader(tic *machines.Tic) error {
	for {
		var msg Msg
		err := cws.ws.conn.ReadJSON(&msg)
		if err != nil {
			log.Fatal(err)
			/*
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			*/
		}

		switch msg.Type {
		case "tic-state":
			tic.WriteImportCode(msg.TicState)
		}
	}
}

// #TODO: fatalErr
func (cws *ClientWS) clientWsWriter(tic *machines.Tic, identity *Identity) {
	err := cws.ws.sendIdentity(identity)
	if err != nil {
		log.Fatal(err)
	}

	fileCheckTicker := time.NewTicker(fileCheckPeriod)
	defer func() {
		fileCheckTicker.Stop()
	}()

	var lastTicState *machines.TicState
	for {
		select {
		//		case <-done:
		//			return
		case <-fileCheckTicker.C:
			// Sends a the local file to the server periodically...
			ticState, err := tic.ReadExportCode()
			if err != nil {
				log.Fatal(err)
				break
			}

			// If we have a previous state and there's no change, don't send
			if lastTicState != nil && lastTicState.IsEqual(*ticState) {
				// Don't send if no change
				break
			}

			err = cws.ws.sendCode(*ticState)
			if err != nil {
				log.Fatal(err)
				break
			}

			lastTicState = ticState
			/*
				case <-interrupt:
					log.Println("interrupt")

						// Cleanly close the connection by sending a close message and then
						// waiting (with timeout) for the server to close the connection.
						err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						if err != nil {
							log.Println("ERR write close:", err)
							return
						}
						select {
						case <-done:
						case <-time.After(time.Second):
						}
					return
			*/
		}
	}
}

// #TODO: (Maybe) Check whether the file has changed before sending
func readFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	return data, nil
}
