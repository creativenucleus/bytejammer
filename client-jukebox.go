package main

import (
	"log"

	"github.com/creativenucleus/bytejammer/comms"
)

func startClientJukebox(host string, port int, playlist *Playlist) error {
	ch := make(chan comms.Msg)
	j, err := NewJukebox(playlist, &ch)
	if err != nil {
		return err
	}

	ws, err := NewWebSocketLink(host, port, "/ws-bytejam")
	if err != nil {
		return err
	}
	defer ws.Close()

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if ok {
					switch msg.Type {
					case "tic-state":
						// #TODO: line endings for data? UTF-8?
						msg := comms.Msg{Type: "tic-state", TicState: msg.TicState}
						err = ws.sendData(msg)
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
