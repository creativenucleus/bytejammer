package main

import (
	"fmt"
	"log"
	"math/rand"
)

func startLocalJukebox(workDir string, playlist *Playlist) error {
	fmt.Printf("Starting local jukebox containing %d items\n", len(playlist.items))

	ch := make(chan Msg)

	j, err := NewJukebox(playlist, &ch)
	if err != nil {
		return err
	}

	slug := fmt.Sprint(rand.Intn(10000))
	tic, err := newServerTic(workDir, slug)
	if err != nil {
		return err
	}
	defer tic.shutdown()

	go func() {
		for {
			select {
			case msg, ok := <-ch:
				if ok {
					switch msg.Type {
					case "code":
						err = tic.importCode(msg.Code)
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
