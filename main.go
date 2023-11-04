package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/creativenucleus/bytejammer/config"
	"github.com/creativenucleus/bytejammer/util"
)

const (
	RELEASE_TITLE = "Appealing Apricot"
)

func main() {
	// Make our working directory
	err := util.EnsurePathExists(config.WORK_DIR, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("==================\n")
	fmt.Printf("Starting ByteJammer\n")
	fmt.Printf("(%s edition)\n", RELEASE_TITLE)
	fmt.Printf("==================\n")

	app := &cli.App{
		DefaultCommand: "jukebox",
		Commands: []*cli.Command{
			{
				Name:  "make-identity",
				Usage: "Create an identity for joining a server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Name to display",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					name := cCtx.String("name")
					err := makeIdentity(name)
					if err != nil {
						log.Fatal(err)
					}

					return nil
				},
			}, {Name: "jukebox",
				Usage: "Run jukebox mode",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "playlist",
						Usage: "Playlist file (empty for LCDZ playlist)",
					},
					&cli.StringFlag{
						Name:  "playtime",
						Usage: "Playtime for each item, default 7 sec",
					},
				},
				Action: func(cCtx *cli.Context) error {
					var_playtime := cCtx.Uint("playtime")
					fmt.Printf("default playtime is %d\n", var_playtime)

					// var playtime = time.Duration(cCtx.Uint("playtime")) * time.Second
					var playtime = time.Duration(var_playtime) * time.Second

					playlistFilename := cCtx.String("playlist")
					playlist, err := readPlaylist(playlistFilename)
					if err != nil {
						log.Fatal(err)
					}
					// err = startLocalJukebox(playlist)
					err = startLocalJukebox(playlist, playtime)
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
						Name:  "localport",
						Usage: "Specify a local port for running the management panel",
						Value: "2000",
					},
					&cli.StringFlag{
						Name:  "broadcast",
						Usage: "Broadcast type (valid: nusan)",
					},
				},
				Action: func(cCtx *cli.Context) error {
					port := cCtx.Int("localport")
					/*
						broadcast := cCtx.String("broadcast")

						var broadcaster *NusanLauncher
						if broadcast != "" {
							if broadcast == "nusan" {
								broadcaster, err = NusanLauncherConnect(4455)
								if err != nil {
									log.Fatal(err)
								}
							} else {
								log.Fatal(errors.New("Unhandled broadcast type"))
							}
						}*/

					err := startHostPanel(port)
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
						Name:  "localport",
						Usage: "Specify a local port for running the management panel",
						Value: "1000",
					},
				},
				Action: func(cCtx *cli.Context) error {
					port := cCtx.Int("localport")

					err = startClientPanel(port)
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
					&cli.StringFlag{
						Name:     "port",
						Usage:    "Port",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "playlist",
						Usage: "Playlist file (empty for LCDZ playlist)",
					},
					&cli.StringFlag{
						Name:  "playtime",
						Usage: "Playtime for each item, default 7 sec",
					},
				},
				Action: func(cCtx *cli.Context) error {
					var_playtime := cCtx.Uint("playtime")
					fmt.Printf("default playtime is %d\n", var_playtime)
					// var playtime = time.Duration(cCtx.Uint("playtime")) * time.Second
					var playtime = time.Duration(var_playtime) * time.Second

					host := cCtx.String("host")
					port := cCtx.Int("port")
					playlistFilename := cCtx.String("playlist")
					playlist, err := readPlaylist(playlistFilename)
					if err != nil {
						log.Fatal(err)
					}

					err = startClientJukebox(host, port, playtime, playlist)
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
