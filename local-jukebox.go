package main

import (
	"fmt"
	"log"

	"github.com/creativenucleus/bytejammer/comms"
	"github.com/creativenucleus/bytejammer/machines"
)

func startLocalJukebox(playlist *Playlist) error {
	fmt.Printf("Starting local jukebox containing %d items\n", len(playlist.items))

	ch := make(chan comms.Msg)

	j, err := NewJukebox(playlist, &ch)
	if err != nil {
		return err
	}

	m, err := machines.LaunchMachine("TIC-80", true, false, false)
	if err != nil {
		return fmt.Errorf("ERR launch machine: %s", err)
	}

	defer m.Shutdown()

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if ok {
					switch msg.Type {
					case "tic-state":
						err = m.Tic.WriteImportCode(msg.TicState.State)
						if err != nil {
							// #TODO: soften!
							log.Fatal(err)
						}
					}
				}
			}
		}
	}()

	j.start()
	for {
	}
}
