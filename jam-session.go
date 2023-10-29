package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/creativenucleus/bytejammer/config"
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/creativenucleus/bytejammer/util"
	"github.com/google/uuid"
)

type JamSession struct {
	port int
	// Our friendly name...
	name string
	// The slug we can use to saved this config to disk...
	slug      string
	startTime time.Time

	// The connections, machines, and the connections between them
	manager *JamSessionManager

	chLog chan string
}

func startJamSession(port int, name string, chLog chan string) (*JamSession, error) {
	nameSlug := util.GetSlug(name)
	if nameSlug == "" {
		return nil, errors.New("Invalid session name - unable to make slug")
	}

	now := time.Now()
	js := JamSession{
		port:      port,
		name:      name,
		slug:      fmt.Sprintf("%s_%s", util.GetSlugFromTime(now), nameSlug),
		startTime: now,
		manager:   makeJamSessionManager(),
		chLog:     chLog,
	}

	basePath := js.getBasePath()
	chLog <- fmt.Sprintf("Creating directory: %s", basePath)
	err := util.EnsurePathExists(basePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	config := getJamSessionConfig(js)
	configData, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(fmt.Sprintf("%s/config.json", basePath), configData, 0644)
	if err != nil {
		return nil, err
	}

	err = js.start()
	if err != nil {
		return nil, err
	}

	return &js, nil
}

func (js *JamSession) start() error {
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

func (js *JamSession) wsBytejam() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		js.chLog <- fmt.Sprintf("Client connected")

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			js.chLog <- fmt.Sprintf("Client connected but couldn't upgrade: %s", err)
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

		jsConn := NewJamSessionConnection(conn)
		js.manager.registerConn(jsConn)

		go jsConn.runServerWsConnRead(js)
		go jsConn.runServerWsConnWrite(js)

		// #TODO: send the server status
		// hp.sendServerStatus(true)

		// #TODO: handle exit
		for {
			select {
			case <-jsConn.signalKick:
				// #TODO: Close down read and write channels
				fmt.Println("KICKED")
				return
			}
		}
	}
}

func (js *JamSession) stop() error {
	fmt.Println("JamSession->stop not yet implemented")
	return nil
}

func (js *JamSession) getBasePath() string {
	return fmt.Sprintf("%sserver-data/%s", config.WORK_DIR, js.slug)
}

func (js *JamSession) getStatus() MsgServerStatus {
	msg := MsgServerStatus{
		Type: "server-status",
	}

	for _, jc := range js.manager.conns {
		status := "waiting"
		machineUuid := ""
		machine := js.manager.getMachineForConn(jc)
		if machine != nil {
			status = fmt.Sprintf("Connected: %s", machine.MachineName)
			machineUuid = machine.Uuid.String()
		}

		msg.Data.Clients = append(msg.Data.Clients, struct {
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

	for _, m := range js.manager.machines {
		name := "(unassigned)"
		clientUuid := ""
		client := js.manager.getConnForMachine(m)
		if client != nil {
			name = client.identity.displayName
			clientUuid = client.connUuid.String()
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

func (js *JamSession) startMachine() (*machines.Machine, error) {
	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return nil, err
	}

	js.manager.registerMachine(m)

	js.chLog <- fmt.Sprintf("TIC-80 Launched: %s", m.MachineName)
	return m, err
}

func (js *JamSession) startMachineForConn(connUuid uuid.UUID) (*machines.Machine, error) {
	conn := js.manager.getConn(connUuid)
	if conn == nil {
		return nil, errors.New("Unable to find conn")
	}

	m, err := machines.LaunchMachine("TIC-80", true, true, false)
	if err != nil {
		return nil, err
	}

	js.manager.registerMachine(m)
	js.manager.linkMachineToConn(m.Uuid, conn.connUuid)

	// TODO: May have identity?
	js.chLog <- fmt.Sprintf("TIC-80 Launched: %s for %s", m.MachineName, conn.connUuid)
	return m, err
}

func (js *JamSession) identifyMachines() {
	count := 0
	for _, c := range js.manager.conns {
		m := js.manager.getMachineForConn(c)
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

func (js *JamSession) closeMachine(data DataCloseMachine) {
	// #TODO: unlink and unregister!

	fmt.Printf("CLOSE: %s\n", data.Uuid)
	err := machines.ShutdownMachine(data.Uuid)
	if err != nil {
		js.chLog <- fmt.Sprintf("ERR shutdown: %s", err)
		return
	}

	js.chLog <- fmt.Sprintf("Machine %s closed", data.Uuid)
}

func (js *JamSession) connectMachineClient(data DataConnectMachineClient) {
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
		js.chLog <- fmt.Sprintf("ERR connect: Could not find Machine ID")
		return
	}

	conn := js.manager.getConn(connUuid)
	if conn == nil {
		js.chLog <- fmt.Sprintf("ERR connect: Could not find Jammer ID")
		return
	}

	js.manager.linkMachineToConn(machineUuid, connUuid)

	js.chLog <- fmt.Sprintf("Connected %s to %s", data.ClientUuid, data.MachineUuid)
}

func (js *JamSession) disconnectMachineClient(data DataDisconnectMachineClient) {
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

	conn := js.manager.getConn(connUuid)
	if conn == nil {
		js.chLog <- fmt.Sprintf("ERR connect: Could not find Jammer ID")
		return
	}

	machine := js.manager.getMachineForConn(conn)
	if machine == nil {
		js.chLog <- fmt.Sprintf("ERR connect: Jammer does not have a machine")
	}

	if machine.Uuid != machineUuid {
		js.chLog <- fmt.Sprintf("ERR connect: Jammer's machine ID does not match the requested one")
	}

	js.manager.unlinkMachineFromConn(machineUuid, connUuid)

	js.chLog <- fmt.Sprintf("Disconnected %s from %s", data.ClientUuid, data.MachineUuid)
}
