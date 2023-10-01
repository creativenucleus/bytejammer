package main

import (
	"fmt"
	"log"
	"time"

	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/machines"
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
		code := machines.TicCodeReplace(embed.LuaJukebox, map[string]string{
			"PLAYLIST_ITEM_COUNT": fmt.Sprintf("%d", len(j.playlist.items)),
			"RELEASE_TITLE":       RELEASE_TITLE,
		})

		(*j.comms) <- Msg{Type: "code", Code: machines.TicCodeAddRunSignal(code)}

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

				code := playlistItem.code
				if playlistItem.author != "" {
					code = machines.TicCodeAddAuthor(code, playlistItem.author)
				}

				code = machines.TicCodeAddRunSignal(code)
				(*j.comms) <- Msg{Type: "code", Code: code}
			}
		}
	}()
}
