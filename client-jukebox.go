package main

import "log"

func startClientJukebox(host string, port int, playlist *Playlist) error {
	ch := make(chan Msg)
	j, err := NewJukebox(playlist, &ch)
	if err != nil {
		return err
	}

	ws, err := NewWebSocketClient(host, port)
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
						err := ws.sendCode(msg.TicState)
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
