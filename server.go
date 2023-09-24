package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type Server struct {
	port     int
	filename string
	server   *http.Server
}

func startServer(workDir string, port int) error {
	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	/*
		fs := http.FileServer(http.Dir("./web-static"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))

		http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/web-static/favicon/favicon.ico")
		})

		fmt.Printf("In a web browser, go to http://localhost:%d/server\n", port)

		http.HandleFunc("/", webIndex)
			http.HandleFunc("/server", webServer)
			http.HandleFunc("/server-ws", wsServer)
	*/

	http.HandleFunc("/ws-bytejam", wsBytejam(workDir))
	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

/*
func webIndex(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(indexHtml)
	if err != nil {
		log.Println("write:", err)
	}
}

func webServer(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(serverHtml)
	if err != nil {
		log.Println("write:", err)
	}
}

func wsServer(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		var msg Msg
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}

		msg = Msg{Type: "ping", Data: []byte("piooong")}
		err = c.WriteJSON(msg)
		if err != nil {
			log.Println("write:", err)
		}
	}
}
*/

func wsBytejam(workDir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("ERR upgrade:", err)
			return
		}
		defer conn.Close()

		slug := fmt.Sprint(rand.Intn(10000))
		tic, err := newServerTic(workDir, slug)
		if err != nil {
			log.Print("ERR new TIC:", err)
			return
		}
		defer tic.shutdown()

		for {
			var msg Msg
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("read:", err)
				break
			}

			switch msg.Type {
			case "code":
				err = tic.importCode(msg.Data)
				if err != nil {
					log.Println("ERR read:", err)
					break
				}
			}

			msg = Msg{Type: "ping", Data: []byte("piooong")}
			err = conn.WriteJSON(msg)
			if err != nil {
				log.Println("ERR write:", err)
			}
		}
	}
}
