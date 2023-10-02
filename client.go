package main

import (
	"time"
)

const (
	fileCheckPeriod = 3 * time.Second
)

func startClient(port int) error {
	err := startClientPanel(port)
	if err != nil {
		return err
	}

	return nil
}
