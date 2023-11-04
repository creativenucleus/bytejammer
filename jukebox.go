package main

import (
	"fmt"
	"log"
	"time"

	"github.com/creativenucleus/bytejammer/comms"
	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/machines"
)

const (
	rotatePeriod = 7 * time.Second
)

type Jukebox struct {
	playlist *Playlist
	playtime time.Duration
	comms    *chan comms.Msg
}

func NewJukebox(playlist *Playlist, playtime time.Duration, comms *chan comms.Msg) (*Jukebox, error) {
	log.Printf("-> Launching Jukebox for playlist")

	j := Jukebox{
		comms:    comms,
		playtime: playtime,
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

		(*j.comms) <- comms.Msg{Type: "tic-state", TicState: tsFirst}

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

				var timeoutduration time.Duration
				//from json file
				if playlistItem.playtime > 0 {
					timeoutduration = time.Duration(playlistItem.playtime) * time.Second
				}
				//from command prompt
				if j.playtime > 0 {
					timeoutduration = time.Duration(j.playtime)
				}
				//internal default
				if timeoutduration == 0 {
					timeoutduration = rotatePeriod
				}
				fmt.Printf("Playtime: %s\n", timeoutduration)
				rotateTicker.Reset(timeoutduration)

				code := playlistItem.code

				/*
					-- Removes vbank 1, so nerfed for now!
					if playlistItem.author != "" {
						code = machines.CodeAddAuthorShim(code, playlistItem.author)
					}
				*/

				ts := machines.MakeTicStateRunning(code)
				(*j.comms) <- comms.Msg{Type: "tic-state", TicState: ts}
			}
		}
	}()
}
