package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer/comms"
	"github.com/creativenucleus/bytejammer/config"
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/creativenucleus/bytejammer/util"
)

type ClientWS struct {
	ws       *WebSocketLink
	chMsg    chan comms.Msg
	basepath string
}

func startClientServerConn(host string, port int, identity *Identity, chServerStatus chan comms.DataClientStatus) error {
	chServerStatus <- comms.DataClientStatus{IsConnected: false}
	cws := ClientWS{
		chMsg: make(chan comms.Msg),
	}

	cws.basepath = filepath.Clean(fmt.Sprintf("%sclient-data/%s", config.WORK_DIR, util.GetSlugFromTime(time.Now())))
	//	chLog <- fmt.Sprintf("Creating directory: %s", cws.basepath)
	err := util.EnsurePathExists(cws.basepath, os.ModePerm)
	if err != nil {
		return err
	}

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
	chServerStatus <- comms.DataClientStatus{IsConnected: true}

	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return err
	}
	defer m.Shutdown()

	// #TODO: shift import / export to *Machine?
	go cws.clientWsReader(m.Tic, identity)
	go cws.clientWsWriter(m.Tic, identity)

	// Lock #TODO: use a channel to escape
	for {
		// Removes 100% CPU warning - but this should really be restructured
		time.Sleep(10 * time.Second)
	}
}

func clientOpenConnection(host string, port int) (*WebSocketLink, error) {
	ws, err := NewWebSocketLink(host, port, "/ws-bytejam")
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (cws *ClientWS) clientWsReader(tic *machines.Tic, identity *Identity) error {
	for {
		var msg comms.Msg
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
		case "challenge-request":
			cws.handleChallengeRequest(msg.ChallengeRequest.Challenge, identity)

		case "tic-state":
			tic.WriteImportCode(msg.TicState.State, true)
		}
	}
}

func (cws *ClientWS) handleChallengeRequest(challenge string, identity *Identity) error {
	data, err := hex.DecodeString(challenge)
	if err != nil {
		return err
	}

	signed, err := identity.Crypto.Sign(data)
	if err != nil {
		return err
	}

	cws.chMsg <- comms.Msg{Type: "challenge-response", ChallengeResponse: comms.DataChallengeResponse{
		Challenge: fmt.Sprintf("%x", signed),
	}}

	return nil
}

// #TODO: fatalErr
func (cws *ClientWS) clientWsWriter(tic *machines.Tic, identity *Identity) {
	// Send Identity...
	publicKeyRaw, err := identity.Crypto.PublicKeyToPem()
	if err != nil {
		log.Fatal(err)
	}

	msg := comms.Msg{
		Type: "identity",
		Identity: comms.DataIdentity{
			Uuid:        identity.Uuid.String(),
			DisplayName: identity.DisplayName,
			PublicKey:   publicKeyRaw,
		},
	}
	err = cws.ws.sendData(&msg)
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
		case msg := <-cws.chMsg:
			cws.ws.sendData(msg)
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

			if ticState.IsRunning {
				err := cws.saveCode(ticState.Code)
				if err != nil {
					log.Fatal(err)
					break
				}
			}

			// #TODO: line endings for data? UTF-8?
			msg := comms.Msg{Type: "tic-state", TicState: comms.DataTicState{
				State: *ticState,
			}}
			err = cws.ws.sendData(msg)
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

func (cws *ClientWS) saveCode(code []byte) error {
	path := filepath.Clean(fmt.Sprintf("%s/code-%s.lua", cws.basepath, util.GetSlugFromTime(time.Now())))
	return os.WriteFile(path, code, 0644)
}
