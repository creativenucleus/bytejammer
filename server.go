package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"

	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/machines"
	"github.com/creativenucleus/bytejammer/server"
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

func startServer(port int, broadcaster *NusanLauncher) error {
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
	http.HandleFunc("/ws-bytejam", wsBytejam(&s, broadcaster))
	http.HandleFunc(fmt.Sprintf("/%s/api/fantasy-machine.json", session), s.apiFantasyMachine)
	if err := webServer.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func webIndex(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(embed.ServerIndexHtml)
	if err != nil {
		log.Println("write:", err)
	}
}

func webOperator(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(embed.ServerOperatorHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func (cp *Server) apiFantasyMachine(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// #TODO: Cleaner way to do this?
		type reqType struct {
			Platform string `json:"platform"`
			Mode     string `json:"mode"`
		}

		var req reqType
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		switch req.Mode {
		case "jammer":
			_, err := machines.LaunchMachine("TIC-80", true, true, false)
			if err != nil {
				// #TODO: propagate error type
				apiOutErr(w, err, http.StatusBadRequest)
				return
			}
		case "jukebox":
			playlist, err := readPlaylist("")
			if err != nil {
				apiOutErr(w, err, http.StatusInternalServerError)
				return
			}

			err = startLocalJukebox(playlist)
			if err != nil {
				apiOutErr(w, err, http.StatusInternalServerError)
				return
			}
		default:
			apiOutErr(w, errors.New("Unexpected mode (should be jammer or jukebox)"), http.StatusBadRequest)
		}

		apiOutResponse(w, nil, http.StatusCreated)

	default:
		apiOutErr(w, errors.New("Method not allowed"), http.StatusMethodNotAllowed)
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
			s.sendServerStatus(c)
		}
	}
}

func wsBytejam(s *Server, broadcaster *NusanLauncher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("ERR upgrade:", err)
			return
		}
		defer conn.Close()

		var m *machines.Machine
		if broadcaster != nil {
			/*			tic, err = machines.NewNusanServerTic(slug, broadcaster)
						if err != nil {
							log.Print("ERR new TIC:", err)
							return
						}
			*/
			log.Print("not implemented")
			return
		} else {
			m, err = machines.LaunchMachine("TIC-80", true, false, true)
			if err != nil {
				log.Print("ERR new TIC:", err)
				return
			}
		}
		defer m.Shutdown()

		// #TODO: Write lock this...
		client := JamClient{
			conn: conn,
			id:   len(s.clients),
		}
		s.clients = append(s.clients, &client)

		go client.runServerWsClientRead(m.Tic)
		//		go runServerWsClientWrite(conn, tic)

		// #TODO: handle exit
		for {
		}
	}
}

func (jc *JamClient) runServerWsClientRead(tic *machines.Tic) {
	for {
		var msg Msg
		err := jc.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		case "code":
			code := msg.Code
			if jc.displayName != "" {
				code = machines.TicCodeAddAuthor(code, jc.displayName)
			}

			err = tic.ImportCode(code)
			if err != nil {
				log.Println("ERR read:", err)
				break
			}
		case "identity":
			jc.displayName = string(msg.Code)
			fmt.Println(jc.displayName)

		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

type MsgServerStatus struct {
	Type string `json:"type"`
	Data struct {
		Clients []struct {
			DisplayName  string
			Status       string
			LastPingTime time.Time
		}
		FantasyMachines []struct {
			MachineName       string
			Platform          string
			Status            string
			JammerDisplayName string
			LastSnapshotTime  time.Time
		}
	} `json:"data"`
}

func (s *Server) sendServerStatus(c *websocket.Conn) {
	msg := MsgServerStatus{
		Type: "server-status",
	}

	for _, jc := range s.clients {
		msg.Data.Clients = append(msg.Data.Clients, struct {
			DisplayName  string
			Status       string
			LastPingTime time.Time
		}{
			DisplayName:  jc.displayName,
			Status:       "waiting",
			LastPingTime: time.Now(),
		})
	}

	for i, m := range machines.MACHINES {
		msg.Data.FantasyMachines = append(msg.Data.FantasyMachines, struct {
			MachineName       string
			Platform          string
			Status            string
			JammerDisplayName string
			LastSnapshotTime  time.Time
		}{
			MachineName:       server.GetFunName(i),
			Platform:          m.Platform,
			Status:            "running",
			JammerDisplayName: "----",
			LastSnapshotTime:  time.Now(),
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

	code := machines.TicCodeAddRunSignal(machines.TicCodeReplace(embed.LuaClient, replacements))
	msg := Msg{Type: "code", Code: code}
	err := jc.conn.WriteJSON(msg)
	if err != nil {
		log.Println("ERR write:", err)
	}
}
