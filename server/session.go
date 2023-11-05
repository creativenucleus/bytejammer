package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/creativenucleus/bytejammer/comms"
	"github.com/creativenucleus/bytejammer/config"
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/creativenucleus/bytejammer/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Session struct {
	port int
	// Our friendly name...
	name string
	// The slug we can use to saved this config to disk...
	slug      string
	startTime time.Time

	// The connections, machines, and the connections between them
	switchboard *Switchboard

	chLog chan string
}

func CreateSession(port int, name string, chLog chan string) (*Session, error) {
	nameSlug := util.GetSlug(name)
	if nameSlug == "" {
		return nil, errors.New("invalid session name - unable to make slug")
	}

	now := time.Now()

	js := Session{
		port:        port,
		name:        name,
		slug:        fmt.Sprintf("%s_%s", util.GetSlugFromTime(now), nameSlug),
		startTime:   now,
		switchboard: makeSwitchboard(),
		chLog:       chLog,
	}

	basePath := js.getBasePath()
	chLog <- fmt.Sprintf("Creating directory: %s", basePath)
	err := util.EnsurePathExists(basePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = js.writeConfig()
	if err != nil {
		return nil, err
	}

	err = js.start()
	if err != nil {
		return nil, err
	}

	// Periodic saving...
	// #TODO: (There must be a bad pattern!)
	go func() {
		for {
			time.Sleep(10 * time.Second)
			err := js.writeConfig()
			if err != nil {
				chLog <- fmt.Sprintf("ERR save config: %s", err)
			}
		}
	}()

	return &js, nil
}

func (js *Session) writeConfig() error {
	config := getSessionConfig(*js)
	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	filepath := fmt.Sprintf("%s/config.json", js.getBasePath())
	err = os.WriteFile(filepath, configData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (js *Session) start() error {
	js.chLog <- fmt.Sprintf("Starting server on port %d", js.port)

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", js.port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	http.HandleFunc("/ws-bytejam", js.wsBytejam())

	// #TODO: catch an error!
	go webServer.ListenAndServe()

	return nil
}

func (js *Session) wsBytejam() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		js.chLog <- "Client connected"

		err := comms.WsUpgrade(w, r, func(conn *websocket.Conn) error {
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

			jsConn := NewJamSessionConnection(conn)
			js.switchboard.registerConn(jsConn)

			go jsConn.runServerWsConnRead(js)
			go jsConn.runServerWsConnWrite(js)

			// #TODO: send the server status
			// hp.sendServerStatus(true)

			// #TODO: handle exit
			for {
				<-jsConn.signalKick
				// #TODO: Close down read and write channels
				fmt.Println("KICKED")
				return nil
			}
		})
		if err != nil {
			js.chLog <- fmt.Sprintf("ws-upgrade: %s", err)
			return
		}
	}
}

func (js *Session) Stop() error {
	fmt.Println("JamSession->stop not yet implemented")
	return nil
}

func (js *Session) getBasePath() string {
	return fmt.Sprintf("%sserver-data/%s", config.WORK_DIR, js.slug)
}

func (js *Session) GetStatus() comms.DataSessionStatus {
	ss := comms.DataSessionStatus{}

	for _, jc := range js.switchboard.conns {
		status := "waiting"
		machineUuid := ""
		machine := js.switchboard.getMachineForConn(jc)
		if machine != nil {
			status = fmt.Sprintf("Connected: %s", machine.MachineName)
			machineUuid = machine.Uuid.String()
		}

		ss.Clients = append(ss.Clients, struct {
			Uuid         string
			DisplayName  string
			ShortUuid    string
			Status       string
			MachineUuid  string
			LastPingTime string
		}{
			Uuid:         jc.connUuid.String(),
			DisplayName:  jc.identity.displayName,
			ShortUuid:    jc.getIdentityShortUuid(),
			Status:       status,
			MachineUuid:  machineUuid,
			LastPingTime: time.Now().Format(time.RFC3339),
		})
	}

	for _, m := range js.switchboard.machines {
		name := "(unassigned)"
		clientUuid := ""
		client := js.switchboard.getConnForMachine(m)
		if client != nil {
			name = client.identity.displayName
			clientUuid = client.connUuid.String()
		}

		ss.Machines = append(ss.Machines, struct {
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

	return ss
}

func (js *Session) StartMachine() (*machines.Machine, error) {
	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return nil, err
	}

	js.switchboard.registerMachine(m)

	js.chLog <- fmt.Sprintf("TIC-80 Launched: %s", m.MachineName)
	return m, err
}

func (js *Session) StartMachineForConn(connUuid uuid.UUID) (*machines.Machine, error) {
	conn := js.switchboard.getConn(connUuid)
	if conn == nil {
		return nil, errors.New("unable to find conn")
	}

	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return nil, err
	}

	js.switchboard.registerMachine(m)
	js.switchboard.linkMachineToConn(m.Uuid, conn.connUuid)

	// TODO: May have identity?
	js.chLog <- fmt.Sprintf("TIC-80 Launched: %s for %s", m.MachineName, conn.connUuid)
	return m, err
}

func (js *Session) IdentifyMachines() {
	count := 0
	for _, c := range js.switchboard.conns {
		m := js.switchboard.getMachineForConn(c)
		if m != nil {
			err := c.sendMachineNameCode(m.MachineName)
			if err != nil {
				js.chLog <- fmt.Sprintln("ERR write:", err)
			}

			count++
		}
	}
	js.chLog <- fmt.Sprintf("Identification sent to %d machines for 30 seconds", count)
}

func (js *Session) CloseMachine(data comms.DataCloseMachine) {
	// #TODO: unlink and unregister!

	fmt.Printf("CLOSE: %s\n", data.Uuid)
	err := machines.ShutdownMachine(data.Uuid)
	if err != nil {
		js.chLog <- fmt.Sprintf("ERR shutdown: %s", err)
		return
	}

	js.chLog <- fmt.Sprintf("Machine %s closed", data.Uuid)
}

func (js *Session) ConnectMachineClient(data comms.DataConnectMachineClient) {
	fmt.Printf("connect: %s to %s\n", data.ClientUuid, data.MachineUuid)

	machineUuid, err := uuid.Parse(data.MachineUuid)
	if err != nil {
		js.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	connUuid, err := uuid.Parse(data.ClientUuid)
	if err != nil {
		js.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	machine := machines.GetMachine(machineUuid)
	if machine == nil {
		js.chLog <- "ERR connect: Could not find Machine ID"
		return
	}

	conn := js.switchboard.getConn(connUuid)
	if conn == nil {
		js.chLog <- "ERR connect: Could not find Jammer ID"
		return
	}

	js.switchboard.linkMachineToConn(machineUuid, connUuid)

	js.chLog <- fmt.Sprintf("Connected %s to %s", data.ClientUuid, data.MachineUuid)
}

func (js *Session) DisconnectMachineClient(data comms.DataDisconnectMachineClient) {
	fmt.Printf("Disconnect: %s to %s\n", data.ClientUuid, data.MachineUuid)

	machineUuid, err := uuid.Parse(data.MachineUuid)
	if err != nil {
		js.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	connUuid, err := uuid.Parse(data.ClientUuid)
	if err != nil {
		js.chLog <- fmt.Sprintf("ERR connect: %s", err)
		return
	}

	conn := js.switchboard.getConn(connUuid)
	if conn == nil {
		js.chLog <- "ERR connect: Could not find Jammer ID"
		return
	}

	machine := js.switchboard.getMachineForConn(conn)
	if machine == nil {
		js.chLog <- "ERR connect: Jammer does not have a machine"
	}

	if machine.Uuid != machineUuid {
		js.chLog <- "ERR connect: Jammer's machine ID does not match the requested one"
	}

	js.switchboard.unlinkMachineFromConn(machineUuid, connUuid)

	js.chLog <- fmt.Sprintf("Disconnected %s from %s", data.ClientUuid, data.MachineUuid)
}
