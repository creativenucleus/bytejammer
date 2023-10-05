package main

import "github.com/creativenucleus/bytejammer/machines"

type MsgTicState struct {
	Code      []byte
	IsRunning bool
	CursorX   int
	CursorY   int
}

type DataCloseMachine struct {
	Uuid string `json:"uuid"`
}

type DataConnectMachineClient struct {
	MachineUuid string `json:"machine-uuid"`
	ClientUuid  string `json:"client-uuid"`
}

type DataDisconnectMachineClient struct {
	MachineUuid string `json:"machine-uuid"`
	ClientUuid  string `json:"client-uuid"`
}

type Msg struct {
	Type                    string                      `json:"type"`
	Identity                string                      `json:"identity,omitempty"`
	TicState                machines.TicState           `json:"tic-state,omitempty"`
	ServerStatus            ClientServerStatus          `json:"server-status,omitempty"`
	ConnectMachineClient    DataConnectMachineClient    `json:"connect-machine-client,omitempty"`
	DisconnectMachineClient DataDisconnectMachineClient `json:"disconnect-machine-client,omitempty"`
	CloseMachine            DataCloseMachine            `json:"close-machine,omitempty"`
}
