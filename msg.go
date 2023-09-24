package main

type Msg struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}
