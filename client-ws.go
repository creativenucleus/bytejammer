package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func startClientServerConn(workDir string, host string, port int, identity *Identity) error {
	var ws *SenderWebSocket
	for {
		// #TODO: This is not the right construction
		var err error
		ws, err = clientOpenConnection(host, port)
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		break
	}
	defer ws.Close()

	slug := fmt.Sprint(rand.Intn(10000))
	tic, err := newClientTic(workDir, slug)
	if err != nil {
		return err
	}
	defer tic.shutdown()

	go clientWsReader(ws, tic)
	go clientWsWriter(ws, tic, identity)

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

func clientWsReader(ws *SenderWebSocket, tic *Tic) error {
	for {
		var msg Msg
		err := ws.conn.ReadJSON(&msg)
		if err != nil {
			log.Fatal(err)
			/*
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			*/
		}

		tic.importCode(msg.Data)
	}
}

// #TODO: fatalErr
func clientWsWriter(ws *SenderWebSocket, tic *Tic, identity *Identity) {
	err := ws.sendIdentity(identity)
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
			data, err := readFile(tic.exportFullpath)
			if err != nil {
				log.Fatal(err)
				break
			}

			if bytes.Equal(lastUpdate, data) {
				// Don't send if no change
				break
			}

			err = ws.sendCode(data)
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
