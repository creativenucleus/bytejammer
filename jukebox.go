package main

import (
	"fmt"
	"log"
	"time"

	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/machines"
)

const (
	rotatePeriod = 15 * time.Second
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
		tsFirst := machines.MakeTicStateRunning(embed.LuaJukebox)
		code := machines.CodeReplace(tsFirst.GetCode(), map[string]string{
			"PLAYLIST_ITEM_COUNT": fmt.Sprintf("%d", len(j.playlist.items)),
			"RELEASE_TITLE":       RELEASE_TITLE,
		})
		tsFirst.SetCode(code)

		(*j.comms) <- Msg{Type: "tic-state", TicState: tsFirst}

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

				/*
					-- Removes vbank 1, so nerfed for now!
					if playlistItem.author != "" {
						code = machines.CodeAddAuthorShim(code, playlistItem.author)
					}
				*/

				ts := machines.MakeTicStateRunning(code)
				(*j.comms) <- Msg{Type: "tic-state", TicState: ts}
			}
		}
	}()
}
