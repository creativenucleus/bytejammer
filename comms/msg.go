package comms

import (
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/creativenucleus/bytejammer/server"
)

type DataLog struct {
	Msg string
}

type MsgTicState struct {
	Code      []byte
	IsRunning bool
	CursorX   int
	CursorY   int
}

type DataClientServerStatus struct {
	IsConnected bool
}

type DataIdentity struct {
	Uuid        string `json:"uuid"`
	DisplayName string `json:"displayName"`
	PublicKey   []byte `json:"publicKey"`
}

type DataTicState machines.TicState

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
	Type string               `json:"type"`
	Data server.SessionStatus `json:"data"`
}

type Msg struct {
	Type                    string                      `json:"type"`
	Identity                DataIdentity                `json:"identity,omitempty"`
	TicState                DataTicState                `json:"tic-state,omitempty"`
	ServerStatus            DataClientServerStatus      `json:"server-status,omitempty"`
	Log                     DataLog                     `json:"log,omitempty"`
	ConnectMachineClient    DataConnectMachineClient    `json:"connect-machine-client,omitempty"`
	DisconnectMachineClient DataDisconnectMachineClient `json:"disconnect-machine-client,omitempty"`
	CloseMachine            DataCloseMachine            `json:"close-machine,omitempty"`
	ChallengeRequest        DataChallengeRequest        `json:"challenge-request,omitempty"`
	ChallengeResponse       DataChallengeResponse       `json:"challenge-response,omitempty"`
}
