package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/creativenucleus/bytejammer/comms"
	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/server"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"
)

const statusSendPeriod = 5 * time.Second

// HostPanel is the web interface for the host to manage their system, and the port should be private to them.
// It handles the startup of a server (potentially multiple)
// It does not handle the connections to the clients directly.

type HostPanel struct {
	wsOperator   *websocket.Conn
	wsMutex      sync.Mutex
	session      *server.Session
	chLog        chan string
	statusTicker *time.Ticker
}

func startHostPanel(port int) error {
	// #TODO: replace
	hostSession := "session"

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	subFs, err := fs.Sub(embed.WebStaticAssets, "web-static")
	if err != nil {
		return err
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFs))))

	fmt.Printf("In a web browser, go to http://localhost:%d/%s/operator\n", port, hostSession)

	hp := HostPanel{
		chLog: make(chan string),
	}

	http.HandleFunc("/", hp.webIndex)
	http.HandleFunc(fmt.Sprintf("/%s/operator", hostSession), hp.webOperator)
	http.HandleFunc(fmt.Sprintf("/%s/ws-operator", hostSession), hp.wsWebOperator())
	http.HandleFunc(fmt.Sprintf("/%s/api/recent-sessions.json", hostSession), hp.webApiRecentSessions)
	http.HandleFunc(fmt.Sprintf("/%s/api/server.json", hostSession), hp.webApiServer)
	http.HandleFunc(fmt.Sprintf("/%s/api/machine.json", hostSession), hp.webApiMachine)
	if err := webServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (hp *HostPanel) webIndex(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(embed.ServerIndexHtml)
	if err != nil {
		log.Println("write:", err)
	}
}

func (hp *HostPanel) webOperator(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(embed.ServerOperatorHtml), w, map[string]stick.Value{
		"release_title": RELEASE_TITLE,
		"session_key":   "session",
	})
	if err != nil {
		log.Println("write:", err)
	}
}

func (hp *HostPanel) wsWebOperator() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		comms.WsUpgrade(w, r, func(conn *websocket.Conn) error {
			hp.wsOperator = conn
			defer func() { hp.wsOperator = nil }()

			go hp.wsOperatorRead()
			go hp.wsOperatorWrite()

			// #TODO: handle exit
			for {
				// Removes 100% CPU warning - but this should really be restructured
				time.Sleep(10 * time.Second)
			}
		})
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
	}
}

func (hp *HostPanel) wsOperatorRead() {
	for {
		var msg comms.Msg
		err := hp.wsOperator.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		case "create-slot":
			hp.handleCreateSlot()
		case "identify-machines":
			hp.handleIdentifyMachines()
		case "connect-machine-client":
			hp.handleConnectMachineClient(msg.ConnectMachineClient)
		case "disconnect-machine-client":
			hp.handleDisconnectMachineClient(msg.DisconnectMachineClient)
		case "close-machine":
			hp.handleCloseMachine(msg.CloseMachine)
		case "stop-server":
			hp.handleStopServer()

		default:
			log.Printf("Message Type not understood: %s\n", msg.Type)
		}
	}
}

func (hp *HostPanel) wsOperatorWrite() {
	hp.statusTicker = time.NewTicker(statusSendPeriod)
	defer func() {
		hp.statusTicker.Stop()
	}()

	for {
		select {
		//		case <-done:
		//			return
		case <-hp.statusTicker.C:
			hp.sendServerStatus(false)

		// Is this in the right place?...
		case logMsg := <-hp.chLog:
			hp.sendLog(logMsg)
		}
	}
}

func (hp *HostPanel) webApiRecentSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		recentSessions, err := server.GetRecentSessions()
		if err != nil {
			apiOutErr(w, err, http.StatusInternalServerError)
			return
		}
		apiOutResponse(w, recentSessions, http.StatusOK)

	default:
		apiOutErr(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (hp *HostPanel) webApiServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// #TODO: Cleaner way to do this?
		type reqType struct {
			Port        string `json:"port"`
			SessionName string `json:"session-name"`
		}

		var req reqType
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		port, err := strconv.Atoi(req.Port)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		// #TODO: This is not great - return some detail
		if hp.session != nil {
			apiOutErr(w, errors.New("server already running"), http.StatusBadRequest)
			return
		}

		hp.session, err = server.CreateSession(port, req.SessionName, hp.chLog)
		if err != nil {
			hp.chLog <- fmt.Sprintf("server failed to launch: %s", err)
			apiOutErr(w, err, http.StatusInternalServerError)
			return
		}

		hp.sendLog("server launched")
		hp.sendServerStatus(true)

	default:
		apiOutErr(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (hp *HostPanel) webApiMachine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// #TODO: Cleaner way to do this?
		type reqType struct {
			Platform   string `json:"platform"`
			Mode       string `json:"mode"`
			ClientUuid string `json:"client-uuid"`
		}

		var req reqType
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		switch req.Mode {
		case "unassigned":
			_, err := hp.session.StartMachine()
			if err != nil {
				apiOutErr(w, fmt.Errorf("TIC-80 Launch (unassigned): %w", err), http.StatusBadRequest)
				return
			}

			hp.sendLog("TIC-80 Launched (unassigned)")

		case "jammer":
			connUuid, err := uuid.Parse(req.ClientUuid)
			if err != nil {
				apiOutErr(w, fmt.Errorf("TIC-80 Launch (jammer): %w", err), http.StatusBadRequest)
				return
			}

			_, err = hp.session.StartMachineForConn(connUuid)
			if err != nil {
				apiOutErr(w, fmt.Errorf("TIC-80 Launch (jammer): %w", err), http.StatusBadRequest)
				return
			}

			hp.sendLog("TIC-80 Launched for (jammer)")

		case "jukebox":
			playlist, err := readPlaylist("")
			if err != nil {
				apiOutErr(w, err, http.StatusInternalServerError)
				return
			}

			err = startLocalJukebox(playlist, time.Duration(JUKEBOX_PLAYTIME_SECS)*time.Second)
			if err != nil {
				apiOutErr(w, err, http.StatusInternalServerError)
				return
			}

			hp.sendLog("TIC-80 Launched for (playlist)")

		default:
			apiOutErr(w, errors.New("unexpected mode (should be jammer or jukebox)"), http.StatusBadRequest)
		}

		hp.sendServerStatus(true)
		apiOutResponse(w, nil, http.StatusCreated)

	default:
		apiOutErr(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}
}
func (hp *HostPanel) handleStopServer() {
	if hp.session == nil {
		hp.chLog <- "Requested server stop, but no server is running"
		return
	}

	hp.session.Stop()
	hp.sendServerStatus(true)
}

func (hp *HostPanel) handleCreateSlot() {
	if hp.session == nil {
		hp.chLog <- "Requested create slot, but no server is running"
		return
	}

	_, err := server.CreateMachineSlot()
	if err != nil {
		hp.chLog <- err.Error()
	}

	hp.sendServerStatus(true)
}

func (hp *HostPanel) handleIdentifyMachines() {
	if hp.session == nil {
		hp.chLog <- "Requested identify machines, but no server is running"
		return
	}

	hp.session.IdentifyMachines()
	hp.sendServerStatus(true)
}

func (hp *HostPanel) handleConnectMachineClient(data comms.DataConnectMachineClient) {
	if hp.session == nil {
		hp.chLog <- "Requested connect, but no server is running"
		return
	}

	hp.session.ConnectMachineClient(data)
	hp.sendServerStatus(true)
}

func (hp *HostPanel) handleDisconnectMachineClient(data comms.DataDisconnectMachineClient) {
	if hp.session == nil {
		hp.chLog <- "Requested disconnect, but no server is running"
		return
	}

	hp.session.DisconnectMachineClient(data)
	hp.sendServerStatus(true)
}

func (hp *HostPanel) handleCloseMachine(data comms.DataCloseMachine) {
	if hp.session == nil {
		hp.chLog <- "Requested close machine, but no server is running"
		return
	}

	hp.session.CloseMachine(data)
	hp.sendServerStatus(true)
}

// #TODO: resetTicker could be improved - we should set that true if the code requests a status send
func (hp *HostPanel) sendServerStatus(resetTicker bool) {
	// #TODO: This is not great - should be driven by the server tick?
	if hp.session == nil {
		return
	}

	if resetTicker {
		hp.statusTicker.Reset(statusSendPeriod)
	}

	msg := comms.Msg{
		Type:          "session-status",
		SessionStatus: hp.session.GetStatus(),
	}

	err := hp.sendData(&msg)
	if err != nil {
		log.Println("read:", err)
	}
}

func (hp *HostPanel) sendLog(message string) {
	msg := comms.Msg{Type: "log", Log: comms.DataLog{Msg: message}}
	fmt.Printf("-> HOST PANEL: %s\n", message)

	err := hp.sendData(&msg)
	if err != nil {
		log.Println("read:", err)
	}
}

func (hp *HostPanel) sendData(data interface{}) error {
	hp.wsMutex.Lock()
	defer hp.wsMutex.Unlock()
	return hp.wsOperator.WriteJSON(data)
}
