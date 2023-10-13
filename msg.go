package main

import "github.com/creativenucleus/bytejammer/machines"

type DataLog struct {
	Msg string
}

type MsgTicState struct {
	Code      []byte
	IsRunning bool
	CursorX   int
	CursorY   int
}

type DataIdentity struct {
	Uuid        string `json:"uuid"`
	DisplayName string `json:"displayName"`
	PublicKey   []byte `json:"publicKey"`
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

type DataChallengeRequest struct {
	Challenge string `json:"challenge"`
}

type DataChallengeResponse struct {
	Challenge string `json:"challenge"`
}

type MsgServerStatus struct {
	Type string `json:"type"`
	Data struct {
		Clients []struct {
			Uuid         string
			DisplayName  string
			ShortUuid    string
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

type Msg struct {
	Type                    string                      `json:"type"`
	Identity                DataIdentity                `json:"identity,omitempty"`
	TicState                machines.TicState           `json:"tic-state,omitempty"`
	ServerStatus            ClientServerStatus          `json:"server-status,omitempty"`
	Log                     DataLog                     `json:"log,omitempty"`
	ConnectMachineClient    DataConnectMachineClient    `json:"connect-machine-client,omitempty"`
	DisconnectMachineClient DataDisconnectMachineClient `json:"disconnect-machine-client,omitempty"`
	CloseMachine            DataCloseMachine            `json:"close-machine,omitempty"`
	ChallengeRequest        DataChallengeRequest        `json:"challenge-request,omitempty"`
	ChallengeResponse       DataChallengeResponse       `json:"challenge-response,omitempty"`
}
