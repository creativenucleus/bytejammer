package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"

	"github.com/creativenucleus/bytejammer/comms"
	"github.com/creativenucleus/bytejammer/embed"
)

// ClientPanel is the web interface for the client to manage their connection and the port should be private to them.
// It handles identity creation, and launching of a connection to the host.
// It does not handle the connection with the host itself.

const (
	fileCheckPeriod = 3 * time.Second
)

type ClientPanel struct {
	// #TODO: lock down to receiver only
	chSendClientStatus chan comms.DataClientStatus
	wsClient           *websocket.Conn
	wsMutex            sync.Mutex
	chLog              chan string
}

func startClientPanel(port int) error {
	// Replace this with a random string...
	session := "session"

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	subFs, err := fs.Sub(embed.WebStaticAssets, "web-static")
	if err != nil {
		return err
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFs))))

	fmt.Printf("In a web browser, go to http://localhost:%d/%s\n", port, session)

	cp := ClientPanel{
		chSendClientStatus: make(chan comms.DataClientStatus),
		chLog:              make(chan string),
	}

	go func() {
		for {
			logMsg := <-cp.chLog
			cp.sendLog(logMsg)
		}
	}()

	http.HandleFunc(fmt.Sprintf("/%s", session), cp.webClientIndex)
	http.HandleFunc(fmt.Sprintf("/%s/api/identity.json", session), cp.webClientApiIdentityJSON)
	http.HandleFunc(fmt.Sprintf("/%s/api/join-server.json", session), cp.webClientApiJoinServerJSON)
	http.HandleFunc(fmt.Sprintf("/%s/ws-client", session), cp.wsWebClient())
	err = webServer.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func (cp *ClientPanel) webClientIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(embed.ClientIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func (cp *ClientPanel) webClientApiIdentityJSON(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		identities, err := getIdentities()
		if err != nil {
			apiOutErr(w, err, http.StatusInternalServerError)
			return
		}

		apiOutResponse(w, identities, http.StatusOK)

	case "POST":
		// #TODO: Cleaner way to do this?
		type reqType struct {
			DisplayName string `json:"displayName"`
		}

		var req reqType
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		err = makeIdentity(req.DisplayName)
		if err != nil {
			apiOutErr(w, err, http.StatusInternalServerError)
			return
		}

		apiOutResponse(w, nil, http.StatusCreated)

	default:
		apiOutErr(w, fmt.Errorf("method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (cp *ClientPanel) webClientApiJoinServerJSON(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		cp.chLog <- "Request: Join Server"

		cp.chSendClientStatus <- comms.DataClientStatus{IsConnected: false}

		// #TODO: Cleaner way to do this?
		type reqType struct {
			Host       string `json:"host"`
			Port       string `json:"port"`
			IdentityId string `json:"identity-id"`
			Message    string `json:"message"`
		}

		var req reqType
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		identity, err := getIdentity(req.IdentityId)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		port, err := strconv.Atoi(req.Port)
		if err != nil {
			apiOutErr(w, err, http.StatusBadRequest)
			return
		}

		err = startClientServerConn(req.Host, port, identity, cp.chSendClientStatus)
		if err != nil {
			apiOutErr(w, err, http.StatusInternalServerError)
			return
		}
		apiOutResponse(w, nil, http.StatusCreated)

	default:
		apiOutErr(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (cp *ClientPanel) wsWebClient() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		comms.WsUpgrade(w, r, func(conn *websocket.Conn) error {
			cp.wsClient = conn
			defer func() { cp.wsClient = nil }()

			go cp.wsRead()
			go cp.wsWrite()

			// #TODO: handle exit
			for {
				// Removes 100% CPU warning - but this should really be restructured
				time.Sleep(10 * time.Second)
			}
		})
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
	}
}

func (cp *ClientPanel) wsRead() {
	for {
		var msg comms.Msg
		err := cp.wsClient.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

func (cp *ClientPanel) wsWrite() {
	/*
		statusTicker := time.NewTicker(statusSendPeriod)
		defer func() {
			statusTicker.Stop()
		}()
	*/
	for {
		status := <-cp.chSendClientStatus
		msg := comms.Msg{Type: "client-status", ClientStatus: status}
		err := cp.sendData(&msg)
		if err != nil {
			// #TODO: relax
			log.Fatal(err)
		}
	}
}

func (cp *ClientPanel) sendLog(message string) {
	msg := comms.Msg{Type: "log", Log: comms.DataLog{Msg: message}}

	err := cp.sendData(&msg)
	if err != nil {
		log.Println("read:", err)
	}
}

func (cp *ClientPanel) sendData(data interface{}) error {
	cp.wsMutex.Lock()
	defer cp.wsMutex.Unlock()
	return cp.wsClient.WriteJSON(data)
}
