package main

import (
	"bytes"
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
		case "code":
			tic.ImportCode(msg.Code)
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

	lastUpdate := []byte{}
	for {
		select {
		//		case <-done:
		//			return
		case <-fileCheckTicker.C:
			// Sends a the local file to the server periodically...
			data, err := readFile(tic.GetExportFullpath())
			if err != nil {
				log.Fatal(err)
				break
			}

			if bytes.Equal(lastUpdate, data) {
				// Don't send if no change
				break
			}

			err = cws.ws.sendCode(data)
			if err != nil {
				log.Fatal(err)
				break
			}

			lastUpdate = data
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
