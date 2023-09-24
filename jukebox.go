package main

import (
	"fmt"
	"log"
	"time"
)

const (
	rotatePeriod = 7 * time.Second
)

type Jukebox struct {
	playlist *Playlist
	comms    *chan Msg
}

func NewJukebox(playlist *Playlist, comms *chan Msg) (*Jukebox, error) {
	log.Printf("-> Launching Jukebox for playlist")

	j := Jukebox{
		comms:    comms,
		playlist: playlist,
	}

	return &j, nil
}

func (j *Jukebox) start() {
	go func() {
		// Send the welcome TIC file
		(*j.comms) <- Msg{Type: "code", Data: ticCodeAddRunSignal(luaWelcome)}

		rotateTicker := time.NewTicker(rotatePeriod)
		defer rotateTicker.Stop()

		for {
			select {
			case <-rotateTicker.C:
				playlistItem, err := j.playlist.getNext()
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

				(*j.comms) <- Msg{Type: "code", Data: ticCodeAddRunSignal(playlistItem.code)}
			}
		}
	}()
}
