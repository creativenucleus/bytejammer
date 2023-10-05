package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"
)

type HostPanel struct {
	wsOperator *websocket.Conn
	wsMutex    sync.Mutex
	server     *Server
	chLog      chan string
}

func startHostPanel(port int) error {
	// #TODO: replace
	session := "session"

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	fs := http.FileServer(http.Dir("./web-static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/web-static/favicon/favicon.ico")
	})

	fmt.Printf("In a web browser, go to http://localhost:%d/%s/operator\n", port, session)

	hp := HostPanel{
		chLog: make(chan string),
	}

	http.HandleFunc("/", hp.webIndex)
	http.HandleFunc(fmt.Sprintf("/%s/operator", session), hp.webOperator)
	http.HandleFunc(fmt.Sprintf("/%s/ws-operator", session), hp.wsWebOperator())
	http.HandleFunc(fmt.Sprintf("/%s/api/server.json", session), hp.webApiServer)
	http.HandleFunc(fmt.Sprintf("/%s/api/machine.json", session), hp.webApiMachine)
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

	err := env.Execute(string(embed.ServerOperatorHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func (hp *HostPanel) wsWebOperator() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		hp.wsOperator, err = wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		defer func() {
			hp.wsOperator.Close()
			hp.wsOperator = nil
		}()

		go hp.wsOperatorRead()
		go hp.wsOperatorWrite()

		// #TODO: handle exit
		for {
		}
	}
}

func (hp *HostPanel) wsOperatorRead() {
	for {
		var msg Msg
		err := hp.wsOperator.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		//		case "reset-clients":
		//			hp.resetAllClients()
		//		case "connect-machine-client":
		//			hp.connectMachineClient(msg.ConnectMachineClient)
		//		case "disconnect-machine-client":
		//			hp.disconnectMachineClient(msg.DisconnectMachineClient)
		//		case "close-machine":
		//			hp.closeMachine(msg.CloseMachine)
		case "stop-server":
			if hp.server == nil {
				// #TODO: log?
				break
			}

			hp.server.stop()

		default:
			log.Printf("Message Type not understood: %s\n", msg.Type)
		}
	}
}

func (hp *HostPanel) wsOperatorWrite() {
	statusTicker := time.NewTicker(statusSendPeriod)
	defer func() {
		statusTicker.Stop()
	}()

	for {
		select {
		//		case <-done:
		//			return
		case <-statusTicker.C:
			hp.sendServerStatus()

		case logMsg := <-hp.chLog:
			hp.sendLog(logMsg)
		}
	}
}

func (hp *HostPanel) webApiServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// #TODO: Cleaner way to do this?
		type reqType struct {
			Port string `json:"port"`
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
		if hp.server != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		hp.server, err = startServer(port, nil, hp.chLog)
		if err != nil {
			apiOutErr(w, err, http.StatusInternalServerError)
			return
		}

		hp.sendLog(fmt.Sprintf("Server launched"))

	default:
		apiOutErr(w, errors.New("Method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (hp *HostPanel) webApiMachine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// #TODO: Cleaner way to do this?
		type reqType struct {
			Platform string `json:"platform"`
			Mode     string `json:"mode"`
		}

		var req reqType
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		switch req.Mode {
		case "jammer":
			m, err := machines.LaunchMachine("TIC-80", true, true, false)
			if err != nil {
				apiOutErr(w, fmt.Errorf("jammer: %w", err), http.StatusBadRequest)
				return
			}

			m.JammerName = "jtruk"
			hp.sendLog(fmt.Sprintf("TIC-80 Launched for %s", m.JammerName))

		case "jukebox":
			hp.sendLog("TIC-80 Launched for (playlist)")

			playlist, err := readPlaylist("")
			if err != nil {
				apiOutErr(w, err, http.StatusInternalServerError)
				return
			}

			err = startLocalJukebox(playlist)
			if err != nil {
				apiOutErr(w, err, http.StatusInternalServerError)
				return
			}
		default:
			apiOutErr(w, errors.New("Unexpected mode (should be jammer or jukebox)"), http.StatusBadRequest)
		}

		apiOutResponse(w, nil, http.StatusCreated)

	default:
		apiOutErr(w, errors.New("Method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (hp *HostPanel) sendServerStatus() {
	// #TODO: This is not great - should be driven by the server tick?
	if hp.server == nil {
		return
	}

	status := hp.server.getStatus()
	err := hp.sendData(&status)
	if err != nil {
		log.Println("read:", err)
	}
}

func (hp *HostPanel) sendLog(message string) {
	msg := MsgLog{Type: "log"}
	msg.Data.Msg = message

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
