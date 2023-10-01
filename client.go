package main

import (
	"time"
)

const (
	fileCheckPeriod = 3 * time.Second
)

func startClient(host string, port int, identity *Identity) error {
	err := startClientPanel(port)
	if err != nil {
		return err
	}

	return nil
}
