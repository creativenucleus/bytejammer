package main

// #TODO: Figure out!

type Msg struct {
	Type         string             `json:"type"`
	Code         []byte             `json:"code,omitempty"`
	ServerStatus ClientServerStatus `json:"server-status,omitempty"`
}
