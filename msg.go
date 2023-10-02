package main

// #TODO: Figure out!

type DataCloseMachine struct {
	Uuid string `json:"uuid"`
}

type Msg struct {
	Type         string             `json:"type"`
	Code         []byte             `json:"code,omitempty"`
	ServerStatus ClientServerStatus `json:"server-status,omitempty"`
	CloseMachine DataCloseMachine   `json:"close-machine,omitempty"`
}
