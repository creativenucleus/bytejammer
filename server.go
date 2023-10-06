package main

import (
	"errors"
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
	links   map[*machines.Machine]*JamClient
	// #TODO: make this work...
	broadcaster *NusanLauncher
}

func MakeServer(chLog chan string) *Server {
	return &Server{
		slug:    getSlugFromTime(time.Now()),
		clients: []*JamClient{},
		chLog:   chLog,

		// #TODO: This feels a bit hacky!
		links: make(map[*machines.Machine]*JamClient),
	}
}

func (s *Server) getClient(clientUuid uuid.UUID) *JamClient {
	for _, c := range s.clients {
		if c.uuid == clientUuid {
			return c
		}
	}
	return nil
}

func (s *Server) attachClientToMachine(machine *machines.Machine, jc *JamClient) {
	// #Mutex?
	s.links[machine] = jc
}

func (s *Server) detachClientFromMachine(machine *machines.Machine) {
	// #Mutex?
	s.links[machine] = nil
}

func (s *Server) getMachineForClient(jc *JamClient) *machines.Machine {
	// #Mutex?
	for machine, itJc := range s.links {
		if itJc == jc {
			return machine
		}
	}

	return nil
}

func (s *Server) getClientForMachine(machine *machines.Machine) *JamClient {
	// #Mutex?
	for itMachine, jc := range s.links {
		if itMachine == machine {
			return jc
		}
	}

	return nil
}

func (s *Server) getJamClient(findUuid uuid.UUID) *JamClient {
	for _, c := range s.clients {
		if c.uuid == findUuid {
			return c
		}
	}

	return nil
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

	s := MakeServer(chLog)

	basePath := s.getBasePath()
	chLog <- fmt.Sprintf("Creating directory: %s", basePath)
	err := ensurePathExists(basePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/ws-bytejam", s.wsBytejam())
	// #TODO: catch an error!
	go webServer.ListenAndServe()

	return s, nil
}

func (s *Server) stop() {
	fmt.Println("#TODO: implement")
}

func (s *Server) wsBytejam() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("ERR upgrade:", err)
			return
		}
		defer conn.Close()

		/*
			var m *machines.Machine
			if s.broadcaster != nil {
							tic, err = machines.NewNusanServerTic(slug, broadcaster)
							if err != nil {
								log.Print("ERR new TIC:", err)
								return
							}

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
		*/

		client := JamClient{
			conn:           conn,
			uuid:           uuid.New(),
			serverBasePath: s.getBasePath(),
		}

		// #TODO: Write lock this...
		s.clients = append(s.clients, &client)

		//		s.attachClientToMachine(m, &client)

		go s.runServerWsClientRead(&client)
		//		go runServerWsClientWrite(conn, tic)

		// #TODO: handle exit
		for {
		}
	}
}

func (s *Server) runServerWsClientRead(jc *JamClient) {
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

			machine := s.getMachineForClient(jc)
			if machine != nil && machine.Tic != nil {
				// Output to Tic
				if jc.displayName != "" {
					ts.SetCode(machines.CodeAddAuthorShim(ts.GetCode(), jc.displayName))
				}

				err = machine.Tic.WriteImportCode(ts)
				if err != nil {
					log.Println("ERR read:", err)
					break
				}
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
			MachineUuid  string
			LastPingTime string
		}
		Machines []struct {
			Uuid              string
			MachineName       string
			ProcessID         int
			Platform          string
			Status            string
			ClientUuid        string
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
		status := "waiting"
		machineUuid := ""
		machine := s.getMachineForClient(jc)
		if machine != nil {
			status = fmt.Sprintf("Connected: %s", machine.MachineName)
			machineUuid = machine.Uuid.String()
		}

		fmt.Println(machineUuid)

		msg.Data.Clients = append(msg.Data.Clients, struct {
			Uuid         string
			DisplayName  string
			Status       string
			MachineUuid  string
			LastPingTime string
		}{
			Uuid:         jc.uuid.String(),
			DisplayName:  jc.displayName,
			Status:       status,
			MachineUuid:  machineUuid,
			LastPingTime: time.Now().Format(time.RFC3339),
		})
	}

	for _, m := range machines.MACHINES {
		name := "(unassigned)"
		clientUuid := ""
		client := s.getClientForMachine(m)
		if client != nil {
			name = client.displayName
			clientUuid = client.uuid.String()
		}

		msg.Data.Machines = append(msg.Data.Machines, struct {
			Uuid              string
			MachineName       string
			ProcessID         int
			Platform          string
			Status            string
			ClientUuid        string
			JammerDisplayName string
			LastSnapshotTime  string
		}{
			Uuid:              m.Uuid.String(),
			MachineName:       m.MachineName,
			ProcessID:         m.Tic.GetProcessID(),
			Platform:          m.Platform,
			Status:            "running",
			ClientUuid:        clientUuid,
			JammerDisplayName: name,
			LastSnapshotTime:  time.Now().Format(time.RFC3339),
		})
	}

	return msg
}

func (s *Server) startMachine() (*machines.Machine, error) {
	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return nil, err
	}

	s.chLog <- fmt.Sprintf("TIC-80 Launched: %s", m.MachineName)

	return m, err
}

func (s *Server) startMachineForClient(clientUuid uuid.UUID) (*machines.Machine, error) {
	client := s.getClient(clientUuid)
	if client == nil {
		return nil, errors.New("Unable to find client")
	}

	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return nil, err
	}

	s.attachClientToMachine(m, s.getJamClient(clientUuid))

	s.chLog <- fmt.Sprintf("TIC-80 Launched: %s for %s", m.MachineName, client.displayName)

	return m, err
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
		s.chLog <- fmt.Sprintf("ERR shutdown: %s", err)
		return
	}

	s.chLog <- fmt.Sprintf("Machine %s closed", data.Uuid)
}

func (s *Server) connectMachineClient(data DataConnectMachineClient) {
	fmt.Printf("connect: %s to %s\n", data.ClientUuid, data.MachineUuid)

	machineUuid, err := uuid.Parse(data.MachineUuid)
	if err != nil {
		s.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	clientUuid, err := uuid.Parse(data.ClientUuid)
	if err != nil {
		s.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	machine := machines.GetMachine(machineUuid)
	if machine == nil {
		s.chLog <- fmt.Sprintf("ERR connect: Could not find Machine ID")
		return
	}

	jamClient := s.getJamClient(clientUuid)
	if jamClient == nil {
		s.chLog <- fmt.Sprintf("ERR connect: Could not find Jammer ID")
		return
	}

	s.attachClientToMachine(machine, jamClient)

	s.chLog <- fmt.Sprintf("Connected %s to %s", data.ClientUuid, data.MachineUuid)
}

func (s *Server) disconnectMachineClient(data DataDisconnectMachineClient) {
	fmt.Printf("Disconnect: %s to %s\n", data.ClientUuid, data.MachineUuid)

	machineUuid, err := uuid.Parse(data.MachineUuid)
	if err != nil {
		s.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	clientUuid, err := uuid.Parse(data.ClientUuid)
	if err != nil {
		s.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	jamClient := s.getJamClient(clientUuid)
	if jamClient == nil {
		s.chLog <- fmt.Sprintf("ERR connect: Could not find Jammer ID")
		return
	}

	machine := s.getMachineForClient(jamClient)
	if machine == nil {
		s.chLog <- fmt.Sprintf("ERR connect: Jammer does not have a machine")
	}

	if machine.Uuid != machineUuid {
		s.chLog <- fmt.Sprintf("ERR connect: Jammer's machine ID does not match the requested one")
	}

	s.detachClientFromMachine(machine)

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
