package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"
)

const statusSendPeriod = 5 * time.Second

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

type JamClient struct {
	conn        *websocket.Conn
	id          int
	displayName string
}

type Server struct {
	//	server   *http.Server
	clients []*JamClient
}

func startServer(workDir string, port int, broadcaster *NusanLauncher) error {
	// Replace this with a random string...
	session := "session"

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	fs := http.FileServer(http.Dir("./web-static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/web-static/favicon/favicon.ico")
	})

	fmt.Printf("In a web browser, go to http://localhost:%d/%s/operator\n", port, session)

	s := Server{
		clients: []*JamClient{},
	}

	http.HandleFunc("/", webIndex)
	http.HandleFunc(fmt.Sprintf("/%s/operator", session), webOperator)
	http.HandleFunc(fmt.Sprintf("/%s/ws-operator", session), wsOperator(&s))
	http.HandleFunc("/ws-bytejam", wsBytejam(&s, workDir, broadcaster))
	if err := webServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func webIndex(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(serverIndexHtml)
	if err != nil {
		log.Println("write:", err)
	}
}

func webOperator(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(serverOperatorHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func wsOperator(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		go wsOperatorRead(s, c)
		go wsOperatorWrite(s, c)

		// #TODO: handle exit
		for {
		}
	}
}

func wsOperatorRead(s *Server, c *websocket.Conn) {
	for {
		var msg Msg
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		case "reset-clients":
			s.resetAllClients()
		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

func wsOperatorWrite(s *Server, c *websocket.Conn) {
	statusTicker := time.NewTicker(statusSendPeriod)
	defer func() {
		statusTicker.Stop()
	}()

	for {
		select {
		//		case <-done:
		//			return
		case <-statusTicker.C:
			s.sendStatus(c)
		}
	}
}

func wsBytejam(s *Server, workDir string, broadcaster *NusanLauncher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("ERR upgrade:", err)
			return
		}
		defer conn.Close()

		slug := fmt.Sprint(rand.Intn(10000))

		var tic *Tic
		if broadcaster != nil {
			tic, err = newNusanServerTic(workDir, slug, broadcaster)
			if err != nil {
				log.Print("ERR new TIC:", err)
				return
			}
		} else {
			tic, err = newServerTic(workDir, slug)
			if err != nil {
				log.Print("ERR new TIC:", err)
				return
			}
		}
		defer tic.shutdown()

		// #TODO: Write lock this...
		client := JamClient{
			conn: conn,
			id:   len(s.clients),
		}
		s.clients = append(s.clients, &client)

		go client.runServerWsClientRead(tic)
		//		go runServerWsClientWrite(conn, tic)

		// #TODO: handle exit
		for {
		}
	}
}

func (jc *JamClient) runServerWsClientRead(tic *Tic) {
	for {
		var msg Msg
		err := jc.conn.ReadJSON(&msg)
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
		case "identity":
			jc.displayName = string(msg.Data)
			fmt.Println(jc.displayName)

		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

type MsgStatus struct {
	Type string `json:"type"`
	Data struct {
		Clients []struct {
			DisplayName string
		}
	} `json:"data"`
}

func (s *Server) sendStatus(c *websocket.Conn) {
	msg := MsgStatus{
		Type: "status",
	}

	for _, jc := range s.clients {
		msg.Data.Clients = append(msg.Data.Clients, struct {
			DisplayName string
		}{
			DisplayName: jc.displayName,
		})
	}

	err := c.WriteJSON(&msg)
	if err != nil {
		log.Println("read:", err)
	}
}

func (s *Server) resetAllClients() {
	fmt.Printf("CLIENTS RESET: %d\n", len(s.clients))
	for _, c := range s.clients {
		c.resetClient()
	}
}

// TODO: Handle error
func (jc *JamClient) resetClient() {
	fmt.Printf("CLIENT RESET: %d\n", jc.id)
	replacements := map[string]string{
		"CLIENT_ID":    fmt.Sprintf("%d", jc.id),
		"DISPLAY_NAME": jc.displayName,
	}

	code := ticCodeAddRunSignal(ticCodeReplace(luaClient, replacements))
	msg := Msg{Type: "code", Data: code}
	err := jc.conn.WriteJSON(msg)
	if err != nil {
		log.Println("ERR write:", err)
	}
}
