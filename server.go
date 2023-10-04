package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"

	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/creativenucleus/bytejammer/server"
)

const statusSendPeriod = 5 * time.Second

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type JamClient struct {
	conn        *websocket.Conn
	uuid        uuid.UUID
	displayName string
}

type Server struct {
	//	server   *http.Server
	clients []*JamClient
	// #TODO: This isn't great - if someone manages to open multiple connections
	wsOperator *websocket.Conn
}

func startServer(port int, broadcaster *NusanLauncher) error {
	// Replace this with a random string...
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

	s := Server{
		clients: []*JamClient{},
	}

	http.HandleFunc("/", webIndex)
	http.HandleFunc(fmt.Sprintf("/%s/operator", session), webOperator)
	http.HandleFunc(fmt.Sprintf("/%s/ws-operator", session), wsOperator(&s))
	http.HandleFunc("/ws-bytejam", wsBytejam(&s, broadcaster))
	http.HandleFunc(fmt.Sprintf("/%s/api/machine.json", session), s.apiMachine)
	if err := webServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func webIndex(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(embed.ServerIndexHtml)
	if err != nil {
		log.Println("write:", err)
	}
}

func webOperator(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(embed.ServerOperatorHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func (cp *Server) apiMachine(w http.ResponseWriter, r *http.Request) {
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
			cp.sendLog(fmt.Sprintf("TIC-80 Launched for %s", m.JammerName))

		case "jukebox":
			cp.sendLog("TIC-80 Launched for (playlist)")

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

func wsOperator(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		s.wsOperator, err = wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		defer func() {
			s.wsOperator.Close()
			s.wsOperator = nil
		}()

		go s.wsOperatorRead()
		go s.wsOperatorWrite()

		// #TODO: handle exit
		for {
		}
	}
}

func (s *Server) wsOperatorRead() {
	for {
		var msg Msg
		err := s.wsOperator.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		case "reset-clients":
			s.resetAllClients()
		case "connect-machine-client":
			s.connectMachineClient(msg.ConnectMachineClient)
		case "disconnect-machine-client":
			s.disconnectMachineClient(msg.DisconnectMachineClient)
		case "close-machine":
			s.closeMachine(msg.CloseMachine)
		default:
			log.Printf("Message Type not understood: %s\n", msg.Type)
		}
	}
}

func (s *Server) wsOperatorWrite() {
	statusTicker := time.NewTicker(statusSendPeriod)
	defer func() {
		statusTicker.Stop()
	}()

	for {
		select {
		//		case <-done:
		//			return
		case <-statusTicker.C:
			s.sendServerStatus()
		}
	}
}

func wsBytejam(s *Server, broadcaster *NusanLauncher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("ERR upgrade:", err)
			return
		}
		defer conn.Close()

		var m *machines.Machine
		if broadcaster != nil {
			/*			tic, err = machines.NewNusanServerTic(slug, broadcaster)
						if err != nil {
							log.Print("ERR new TIC:", err)
							return
						}
			*/
			log.Print("not implemented")
			return
		} else {
			m, err = machines.LaunchMachine("TIC-80", true, false, true)
			if err != nil {
				log.Print("ERR new TIC:", err)
				return
			}

			s.sendLog("TIC-80 Launched")
		}
		defer m.Shutdown()

		// #TODO: Write lock this...
		client := JamClient{
			conn: conn,
			uuid: uuid.New(),
		}
		s.clients = append(s.clients, &client)

		go client.runServerWsClientRead(m.Tic)
		//		go runServerWsClientWrite(conn, tic)

		// #TODO: handle exit
		for {
		}
	}
}

func (jc *JamClient) runServerWsClientRead(tic *machines.Tic) {
	for {
		var msg Msg
		err := jc.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		case "tic-state":
			ts := msg.TicState
			if jc.displayName != "" {
				ts.SetCode(machines.CodeAddAuthorShim(ts.GetCode(), jc.displayName))
			}

			err = tic.WriteImportCode(ts)
			if err != nil {
				log.Println("ERR read:", err)
				break
			}
		case "identity":
			jc.displayName = string(msg.Identity)
			fmt.Println(jc.displayName)

		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

type MsgLog struct {
	Type string `json:"type"`
	Data struct {
		Msg string
	} `json:"data"`
}

func (s *Server) sendLog(message string) {
	msg := MsgLog{Type: "log"}
	msg.Data.Msg = message

	err := s.wsOperator.WriteJSON(&msg)
	if err != nil {
		log.Println("read:", err)
	}
}

type MsgServerStatus struct {
	Type string `json:"type"`
	Data struct {
		Clients []struct {
			Uuid         string
			DisplayName  string
			Status       string
			LastPingTime string
		}
		Machines []struct {
			Uuid              string
			MachineName       string
			ProcessID         int
			Platform          string
			Status            string
			JammerDisplayName string
			LastSnapshotTime  string
		}
	} `json:"data"`
}

func (s *Server) sendServerStatus() {
	msg := MsgServerStatus{
		Type: "server-status",
	}

	for _, jc := range s.clients {
		msg.Data.Clients = append(msg.Data.Clients, struct {
			Uuid         string
			DisplayName  string
			Status       string
			LastPingTime string
		}{
			Uuid:         jc.uuid.String(),
			DisplayName:  jc.displayName,
			Status:       "waiting",
			LastPingTime: time.Now().Format(time.RFC3339),
		})
	}

	for i, m := range machines.MACHINES {
		msg.Data.Machines = append(msg.Data.Machines, struct {
			Uuid              string
			MachineName       string
			ProcessID         int
			Platform          string
			Status            string
			JammerDisplayName string
			LastSnapshotTime  string
		}{
			Uuid:              m.Uuid.String(),
			MachineName:       server.GetFunName(i),
			ProcessID:         m.Tic.GetProcessID(),
			Platform:          m.Platform,
			Status:            "running",
			JammerDisplayName: m.JammerName,
			LastSnapshotTime:  time.Now().Format(time.RFC3339),
		})
	}

	err := s.wsOperator.WriteJSON(&msg)
	if err != nil {
		log.Println("read:", err)
	}
}

func (s *Server) resetAllClients() {
	fmt.Printf("CLIENTS RESET: %d\n", len(s.clients))
	for _, c := range s.clients {
		c.resetClient()
	}
}

func (s *Server) closeMachine(data DataCloseMachine) {
	fmt.Printf("CLOSE: %s\n", data.Uuid)
	err := machines.ShutdownMachine(data.Uuid)
	if err != nil {
		log.Println("ERR shutdown:", err)
	}

	s.sendLog(fmt.Sprintf("Machine %s closed", data.Uuid))
}

func (s *Server) connectMachineClient(data DataConnectMachineClient) {
	fmt.Printf("connect: %s to %s\n", data.ClientUuid, data.MachineUuid)
	/*	err := machines.ShutdownMachine(data.Uuid)
		if err != nil {
			log.Println("ERR shutdown:", err)
		}
	*/
	s.sendLog(fmt.Sprintf("Connected %s to %s", data.ClientUuid, data.MachineUuid))
}

func (s *Server) disconnectMachineClient(data DataDisconnectMachineClient) {
	fmt.Printf("Disconnect: %s to %s\n", data.ClientUuid, data.MachineUuid)
	/*	err := machines.ShutdownMachine(data.Uuid)
		if err != nil {
			log.Println("ERR shutdown:", err)
		}
	*/
	s.sendLog(fmt.Sprintf("Disconnected %s from %s", data.ClientUuid, data.MachineUuid))
}

// TODO: Handle error
func (jc *JamClient) resetClient() {
	fmt.Printf("CLIENT RESET: %d\n", jc.uuid)

	ts := machines.MakeTicStateRunning(embed.LuaClient)
	code := machines.CodeReplace(ts.GetCode(), map[string]string{
		"CLIENT_ID":    fmt.Sprintf("%s", "Fake machine name"),
		"DISPLAY_NAME": jc.displayName,
	})
	ts.SetCode(code)

	msg := Msg{Type: "tic-state", TicState: ts}
	err := jc.conn.WriteJSON(msg)
	if err != nil {
		log.Println("ERR write:", err)
	}
}
