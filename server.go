package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/creativenucleus/bytejammer/config"
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
	conn           *websocket.Conn
	wsMutex        sync.Mutex
	uuid           uuid.UUID
	displayName    string
	lastTicState   *machines.TicState
	serverBasePath string
}

type Server struct {
	slug    string
	clients []*JamClient
	chLog   chan string
}

func (s *Server) getBasePath() string {
	return fmt.Sprintf("%sserver-session/%s", config.WORK_DIR, s.slug)
}

func startServer(port int, broadcaster *NusanLauncher, chLog chan string) (*Server, error) {
	chLog <- fmt.Sprintf("Starting server on port %d", port)
	// Replace this with a random string...
	//	session := "session"

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	s := Server{
		slug:    getSlugFromTime(time.Now()),
		clients: []*JamClient{},
		chLog:   chLog,
	}

	basePath := s.getBasePath()
	chLog <- fmt.Sprintf("Creating directory: %s", basePath)
	err := ensurePathExists(basePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/ws-bytejam", wsBytejam(&s, broadcaster))
	// #TODO: catch an error!
	go webServer.ListenAndServe()

	return &s, nil
}

func (s *Server) stop() {
	fmt.Println("#TODO: implement")
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

			s.chLog <- "TIC-80 Launched"
		}
		defer m.Shutdown()

		client := JamClient{
			conn:           conn,
			uuid:           uuid.New(),
			serverBasePath: s.getBasePath(),
		}
		// #TODO: Write lock this...
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

			if jc.lastTicState != nil && ts.IsEqual(*jc.lastTicState) {
				// We already sent this state
				continue
			}

			if ts.IsRunning {
				// #TODO: I don't think this fully works? Seems to save more than it should
				// #TODO: slugify displayName!
				path := fmt.Sprintf("%s/code-%s-%s.lua", jc.serverBasePath, jc.displayName, getSlugFromTime(time.Now()))
				os.WriteFile(path, []byte(ts.GetCode()), 0644)
			}

			if jc.displayName != "" {
				ts.SetCode(machines.CodeAddAuthorShim(ts.GetCode(), jc.displayName))
			}

			err = tic.WriteImportCode(ts)
			if err != nil {
				log.Println("ERR read:", err)
				break
			}

			jc.lastTicState = &ts

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

func (s *Server) getStatus() MsgServerStatus {
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

	return msg
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

	s.chLog <- fmt.Sprintf("Machine %s closed", data.Uuid)
}

func (s *Server) connectMachineClient(data DataConnectMachineClient) {
	fmt.Printf("connect: %s to %s\n", data.ClientUuid, data.MachineUuid)
	/*	err := machines.ShutdownMachine(data.Uuid)
		if err != nil {
			log.Println("ERR shutdown:", err)
		}
	*/
	s.chLog <- fmt.Sprintf("Connected %s to %s", data.ClientUuid, data.MachineUuid)
}

func (s *Server) disconnectMachineClient(data DataDisconnectMachineClient) {
	fmt.Printf("Disconnect: %s to %s\n", data.ClientUuid, data.MachineUuid)
	/*	err := machines.ShutdownMachine(data.Uuid)
		if err != nil {
			log.Println("ERR shutdown:", err)
		}
	*/
	s.chLog <- fmt.Sprintf("Disconnected %s from %s", data.ClientUuid, data.MachineUuid)
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
	err := jc.sendData(msg)
	if err != nil {
		log.Println("ERR write:", err)
	}
}

func (jc *JamClient) sendData(data interface{}) error {
	jc.wsMutex.Lock()
	defer jc.wsMutex.Unlock()
	return jc.conn.WriteJSON(data)
}
