package comms

import (
	"github.com/creativenucleus/bytejammer/machines"
)

type DataLog struct {
	Msg string
}

type DataClientStatus struct {
	IsConnected bool
}

type DataIdentity struct {
	Uuid        string `json:"uuid"`
	DisplayName string `json:"displayName"`
	PublicKey   []byte `json:"publicKey"`
}

type DataTicState struct {
	State machines.TicState
}

type DataCloseMachine struct {
	Uuid string `json:"uuid"`
}

type DataMachineSetup struct {
	SlotID     string `json:"slot-id"`
	Connection string `json:"connection"`
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

type DataSessionStatus struct {
	Port    int
	Clients []struct {
		Uuid         string
		DisplayName  string
		ShortUuid    string
		Status       string
		MachineUuid  string
		LastPingTime string
	}
	Slots []struct {
		Id int
		// #TODO: Fill... (take over from the Machine slice)
		Status            string
		MachineName       string
		ProcessID         int
		Platform          string
		JammerDisplayName string
		LastSnapshotTime  string
		/*
			Uuid              string
			ClientUuid        string
		*/
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
}

type Msg struct {
	Type                    string                      `json:"type"`
	Identity                DataIdentity                `json:"identity,omitempty"`
	TicState                DataTicState                `json:"tic-state,omitempty"`
	ClientStatus            DataClientStatus            `json:"client-status,omitempty"`
	SessionStatus           DataSessionStatus           `json:"session-status,omitempty"`
	Log                     DataLog                     `json:"log,omitempty"`
	ConnectMachineClient    DataConnectMachineClient    `json:"connect-machine-client,omitempty"`
	DisconnectMachineClient DataDisconnectMachineClient `json:"disconnect-machine-client,omitempty"`
	DataMachineSetup        DataMachineSetup            `json:"machine-setup,omitempty"`
	CloseMachine            DataCloseMachine            `json:"close-machine,omitempty"`
	ChallengeRequest        DataChallengeRequest        `json:"challenge-request,omitempty"`
	ChallengeResponse       DataChallengeResponse       `json:"challenge-response,omitempty"`
}
