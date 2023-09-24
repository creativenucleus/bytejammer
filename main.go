package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

// "./playlist/nanogems-test-selection.zip"

func main() {
	workDir := "./work/"

	// Make our working directory
	err := os.MkdirAll(filepath.Clean(workDir), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		DefaultCommand: "jukebox",
		Commands: []*cli.Command{
			{
				Name:  "jukebox",
				Usage: "Run jukebox mode",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "playlist",
						Usage: "Playlist file (empty for LCDZ playlist)",
					},
				},
				Action: func(cCtx *cli.Context) error {
					playlistFilename := cCtx.String("playlist")
					playlist, err := readPlaylist(playlistFilename)
					if err != nil {
						log.Fatal(err)
					}

					err = startStandaloneJukebox(workDir, playlist)
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			}, {
				Name:  "server",
				Usage: "Run server mode",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "host",
						Usage:    "Host",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "port",
						Usage:    "Port",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					playlistFilename := cCtx.String("playlist")
					playlist, err := readPlaylist(playlistFilename)
					if err != nil {
						log.Fatal(err)
					}

					err = startStandaloneJukebox(workDir, playlist)
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			}, {
				Name:  "client",
				Usage: "Run client",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "host",
						Usage:    "Host",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					playlistFilename := cCtx.String("playlist")
					playlist, err := readPlaylist(playlistFilename)
					if err != nil {
						log.Fatal(err)
					}

					err = startStandaloneJukebox(workDir, playlist)
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			}, {
				Name:  "client-jukebox",
				Usage: "Run client-jukebox",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "host",
						Usage:    "Host",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					playlistFilename := cCtx.String("playlist")
					playlist, err := readPlaylist(playlistFilename)
					if err != nil {
						log.Fatal(err)
					}

					err = startStandaloneJukebox(workDir, playlist)
					if err != nil {
						log.Fatal(err)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func readPlaylist(filename string) (*Playlist, error) {
	if filename == "" {
		return NewPlaylistLCDZ()
	}

	fullFilepath := filepath.Clean(filename)
	ext := filepath.Ext(fullFilepath)
	if ext == ".zip" {
		return NewPlaylistFromZip(fullFilepath)
	}

	playlistData, err := os.ReadFile(fullFilepath)
	if err != nil {
		return nil, err
	}

	return NewPlaylistFromJSON(playlistData)
}

func startStandaloneJukebox(workDir string, playlist *Playlist) error {
	ch := make(chan Msg)

	c, err := NewClientJukebox(playlist, &ch)
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
						err = tic.importCode(msg.Data)
						if err != nil {
							// #TODO: soften!
							log.Fatal(err)
						}
					}
				}
			}
		}
	}()

	c.start()
	for {
	}
}
