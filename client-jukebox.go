package main

import (
	"fmt"
	"log"
	"time"
)

const (
	rotatePeriod = 7 * time.Second
)

type ClientJukebox struct {
	playlist *Playlist
	comms    *chan Msg
}

func NewClientJukebox(playlist *Playlist, comms *chan Msg) (*ClientJukebox, error) {
	log.Printf("-> Launching Jukebox for playlist")

	c := ClientJukebox{
		comms:    comms,
		playlist: playlist,
	}

	return &c, nil
}

func (c *ClientJukebox) start() {
	go func() {
		// Send the welcome TIC file
		(*c.comms) <- Msg{Type: "code", Data: ticCodeAddRunSignal(luaWelcome)}

		rotateTicker := time.NewTicker(rotatePeriod)
		defer rotateTicker.Stop()

		for {
			select {
			case <-rotateTicker.C:
				playlistItem, err := c.playlist.getNext()
				if err != nil {
					log.Println("ERR get code:", err)
					break
				}

				fmt.Printf("Playing (TIC):\nLocation: %s\n", playlistItem.location)
				if playlistItem.author != "" {
					fmt.Printf("Author: %s\n", playlistItem.author)
				}

				if playlistItem.description != "" {
					fmt.Printf("Description: %s\n", playlistItem.description)
				}

				(*c.comms) <- Msg{Type: "code", Data: ticCodeAddRunSignal(playlistItem.code)}
			}
		}
	}()
}
