package main

import (
	"fmt"
	"log"

	"github.com/creativenucleus/bytejammer/machines"
)

func startLocalJukebox(playlist *Playlist) error {
	fmt.Printf("Starting local jukebox containing %d items\n", len(playlist.items))

	ch := make(chan Msg)

	j, err := NewJukebox(playlist, &ch)
	if err != nil {
		return err
	}

	m, err := machines.LaunchMachine("TIC-80", true, false, false)
	m.JammerName = "(jukebox)"
	if err != nil {
		return err
	}
	defer m.Shutdown()

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if ok {
					switch msg.Type {
					case "code":
						err = m.Tic.ImportCode(msg.Code)
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
