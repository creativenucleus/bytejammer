package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const statusSendPeriod = 10 * time.Second

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

type ClientConnForServer struct {
	conn *websocket.Conn
	id   int
}

type Server struct {
	//	server   *http.Server
	clients []*ClientConnForServer
}

func startServer(workDir string, port int) error {
	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	fs := http.FileServer(http.Dir("./web-static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/web-static/favicon/favicon.ico")
	})

	fmt.Printf("In a web browser, go to http://localhost:%d/operator\n", port)

	s := Server{
		clients: []*ClientConnForServer{},
	}

	http.HandleFunc("/", webIndex)
	http.HandleFunc("/operator", webOperator)
	http.HandleFunc("/ws-operator", wsOperator(&s))
	http.HandleFunc("/ws-bytejam", wsBytejam(&s, workDir))
	if err := webServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func webIndex(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(indexHtml)
	if err != nil {
		log.Println("write:", err)
	}
}

func webOperator(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(operatorHtml)
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

func wsBytejam(s *Server, workDir string) func(http.ResponseWriter, *http.Request) {
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

		// #TODO: Write lock this...
		s.clients = append(s.clients, &ClientConnForServer{conn: conn, id: len(s.clients)})

		go runServerWsClientRead(conn, tic)
		//		go runServerWsClientWrite(conn, tic)

		// #TODO: handle exit
		for {
		}
	}
}

func runServerWsClientRead(conn *websocket.Conn, tic *Tic) {
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
	}
}

type MsgStatus struct {
	ClientCount int `json:"client-count"`
}

func (s *Server) sendStatus(c *websocket.Conn) {
	data, err := json.Marshal(MsgStatus{len(s.clients)})
	if err != nil {
		log.Println("ERR marshal:", err)
		return
	}

	msg := Msg{
		Type: "status",
		Data: data,
	}

	err = c.WriteJSON(&msg)
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
func (c *ClientConnForServer) resetClient() {
	fmt.Printf("CLIENT RESET: %d\n", c.id)
	replacements := map[string]string{"CLIENT": fmt.Sprintf("%d", c.id)}
	code := ticCodeAddRunSignal(ticCodeReplace(luaClient, replacements))
	msg := Msg{Type: "code", Data: code}
	err := c.conn.WriteJSON(msg)
	if err != nil {
		log.Println("ERR write:", err)
	}
}