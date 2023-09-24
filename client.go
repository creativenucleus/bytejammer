package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const (
	fileCheckPeriod = 3 * time.Second
)

func startClient(workDir string, host string, port int) error {
	ws, err := NewWebSocketClient(host, port)
	if err != nil {
		return err
	}
	defer ws.Close()

	slug := fmt.Sprint(rand.Intn(10000))
	tic, err := newClientTic(workDir, slug)
	if err != nil {
		return err
	}
	defer tic.shutdown()

	fileCheckTicker := time.NewTicker(fileCheckPeriod)
	defer func() {
		fileCheckTicker.Stop()
	}()

	for {
		select {
		//		case <-done:
		//			return
		case <-fileCheckTicker.C:
			data, err := readFile(tic.exportFilename)
			if err != nil {
				return err
			}

			err = ws.sendCode(data)
			if err != nil {
				return err
			}
			/*
				case <-interrupt:
					log.Println("interrupt")

						// Cleanly close the connection by sending a close message and then
						// waiting (with timeout) for the server to close the connection.
						err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						if err != nil {
							log.Println("ERR write close:", err)
							return
						}
						select {
						case <-done:
						case <-time.After(time.Second):
						}
					return
			*/
		}
	}
}

/*
func readFileIfModified(filename string, lastMod time.Time) ([]byte, time.Time, error) {
	fmt.Println("File check")
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, lastMod, err
	}
	if !fi.ModTime().After(lastMod) {
		return nil, lastMod, nil
	}
	p, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, fi.ModTime(), err
	}
	fmt.Println("File READ")
	return p, fi.ModTime(), nil
}
*/

func readFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	return data, nil
}
