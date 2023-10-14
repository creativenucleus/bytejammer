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
	chSendServerStatus chan ClientServerStatus
	wsClient           *websocket.Conn
	wsMutex            sync.Mutex
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
		chSendServerStatus: make(chan ClientServerStatus),
	}
	http.HandleFunc(fmt.Sprintf("/%s", session), cp.webClientIndex)
	http.HandleFunc(fmt.Sprintf("/%s/api/identity.json", session), cp.webClientApiIdentityJSON)
	http.HandleFunc(fmt.Sprintf("/%s/api/join-server.json", session), cp.webClientApiJoinServerJSON)
	http.HandleFunc(fmt.Sprintf("/%s/ws-client", session), cp.wsWebClient())
	if err := webServer.ListenAndServe(); err != nil {
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
		apiOutErr(w, errors.New("Method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (cp *ClientPanel) webClientApiJoinServerJSON(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		cp.chSendServerStatus <- ClientServerStatus{isConnected: false}

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

		err = startClientServerConn(req.Host, port, identity, cp.chSendServerStatus)
		if err != nil {
			apiOutErr(w, err, http.StatusInternalServerError)
			return
		}
		apiOutResponse(w, nil, http.StatusCreated)

	default:
		apiOutErr(w, errors.New("Method not allowed"), http.StatusMethodNotAllowed)
	}
}

func (cp *ClientPanel) wsWebClient() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		cp.wsClient, err = wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer cp.wsClient.Close()

		go cp.wsRead()
		go cp.wsWrite()

		// #TODO: handle exit
		for {
		}
	}
}

func (cp *ClientPanel) wsRead() {
	for {
		var msg Msg
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
		select {
		//		case <-done:
		//			return
		//		case <-statusTicker.C:
		//			fmt.Println("TICKER!")

		case status := <-cp.chSendServerStatus:
			msg := Msg{Type: "server-status", ServerStatus: status}
			err := cp.sendData(&msg)
			if err != nil {
				// #TODO: relax
				log.Fatal(err)
			}
		}
	}
}

func (cp *ClientPanel) sendData(data interface{}) error {
	cp.wsMutex.Lock()
	defer cp.wsMutex.Unlock()
	return cp.wsClient.WriteJSON(data)
}
