package server

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/creativenucleus/bytejammer/comms"
	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/creativenucleus/bytejammer/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// JamSessionConn is a connection to the JamServer.
// It may not not yet be validated, or have an identity.
type SessionConn struct {
	conn           *websocket.Conn
	connUuid       uuid.UUID
	wsMutex        sync.Mutex
	identity       *SessionConnIdentity
	lastTicState   *machines.TicState
	serverBasePath string
	publicKey      []byte // This should be a public key type, and we should manage the challenge status
	signalKick     chan bool
}

type SessionConnIdentity struct {
	uuid        uuid.UUID
	displayName string
	publicKey   []byte
	isConfirmed bool
}

func NewJamSessionConnection(conn *websocket.Conn) *SessionConn {
	client := SessionConn{
		conn:       conn,
		connUuid:   uuid.New(),
		signalKick: make(chan bool),
	}
	return &client
}

// #TODO: make better
func (jc *SessionConn) getIdentityShortUuid() string {
	if jc.identity == nil {
		return "(unknown)"
	}

	return jc.identity.uuid.String()[0:8]
}

func (jc *SessionConn) runServerWsConnRead(js *Session) {
	for {
		var msg comms.Msg
		err := jc.conn.ReadJSON(&msg)
		if err != nil {
			js.chLog <- fmt.Sprintln("read:", err)
			break
		}

		switch msg.Type {
		case "identity":
			identityUuid, err := uuid.Parse(msg.Identity.Uuid)
			if err != nil {
				js.chLog <- fmt.Sprintln("read:", err)
				break
			}

			// #TODO: NB This identity has not yet been challenged
			jc.identity = &SessionConnIdentity{
				uuid:        identityUuid,
				displayName: msg.Identity.DisplayName,
				publicKey:   msg.Identity.PublicKey,
				isConfirmed: false,
			}

			// See whether the identity / public key matches our known one
			// #TODO: something

			// #TODO: Refactor this placeholder!
			// Kick an existing connection off if it has the same identity
			for _, c := range js.switchboard.conns {
				if c != jc && c.identity != nil && c.identity.uuid.String() == jc.identity.uuid.String() {
					c.signalKick <- true
					js.switchboard.unregisterConn(c)
				}
			}

			// Send the challenge
			msg := comms.Msg{Type: "challenge-request", ChallengeRequest: comms.DataChallengeRequest{Challenge: "This will be a random string!"}}
			err = jc.sendData(msg)
			if err != nil {
				js.chLog <- fmt.Sprintln("write:", err)
			}

		case "challenge-response":
			// #TODO: Match identity to any existing (fallen) clients, and stitch together
			// Do two live clients match identity uuid?! What now?
			// Check public key matches known one

			fmt.Println(msg.ChallengeResponse.Challenge)

		case "tic-state":
			ts := msg.TicState.State

			if jc.lastTicState != nil && ts.IsEqual(*jc.lastTicState) {
				// We already sent this state
				continue
			}

			if ts.IsRunning {
				// #TODO: I don't think this fully works? Seems to save more than it should
				// #TODO: slugify displayName!
				path := fmt.Sprintf("%s/code-%s-%s.lua", jc.serverBasePath, jc.identity.displayName, util.GetSlugFromTime(time.Now()))
				os.WriteFile(path, []byte(ts.GetCode()), 0644)
			}

			machine := js.switchboard.getMachineForConn(jc)
			if machine != nil && machine.Tic != nil {
				// Output to Tic
				// Don't shim for now...
				/*
					if jc.identity.displayName != "" {
						ts.SetCode(machines.CodeAddAuthorShim(ts.GetCode(), jc.identity.displayName))
					}
				*/

				err = machine.Tic.WriteImportCode(ts)
				if err != nil {
					js.chLog <- fmt.Sprintln("read:", err)
					break
				}
			}

			jc.lastTicState = &ts

		default:
			js.chLog <- fmt.Sprintf("message not understood: %s", msg.Type)
		}
	}
}

func (jc *SessionConn) runServerWsConnWrite(js *Session) {
	for {
		select {}
	}
}

// TODO: Handle error
func (js *SessionConn) sendMachineNameCode(machineName string) error {
	fmt.Printf("CLIENT RESET: %d\n", js.connUuid)

	ts := machines.MakeTicStateRunning(embed.LuaClient)
	code := machines.CodeReplace(ts.GetCode(), map[string]string{
		"CLIENT_ID":    machineName,
		"DISPLAY_NAME": js.identity.displayName,
	})
	ts.SetCode(code)

	msg := comms.Msg{Type: "tic-state", TicState: comms.DataTicState{
		State: ts,
	}}
	err := js.sendData(msg)
	return err
}

func (jc *SessionConn) sendData(data interface{}) error {
	jc.wsMutex.Lock()
	defer jc.wsMutex.Unlock()
	return jc.conn.WriteJSON(data)
}
