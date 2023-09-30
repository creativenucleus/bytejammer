package main

import (
	"time"
)

const (
	fileCheckPeriod = 3 * time.Second
)

func startClient(workDir string, host string, port int, identity *Identity) error {
	err := startClientPanel(workDir, port)
	if err != nil {
		return err
	}

	return nil
}
