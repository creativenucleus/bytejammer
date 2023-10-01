package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"

	"github.com/creativenucleus/bytejammer/embed"
)

type ClientPanel struct {
	// #TODO: lock down to receiver only
	chSendServerStatus chan ClientServerStatus
}

func startClientPanel(port int) error {
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

	fmt.Printf("In a web browser, go to http://localhost:%d/%s\n", port, session)

	cp := ClientPanel{
		chSendServerStatus: make(chan ClientServerStatus),
	}
	http.HandleFunc(fmt.Sprintf("/%s", session), cp.webClientIndex)
	http.HandleFunc(fmt.Sprintf("/%s/api/identity.json", session), cp.webClientApiIdentityJSON)
	http.HandleFunc(fmt.Sprintf("/%s/api/join-server.json", session), cp.webClientApiJoinServerJSON)
	http.HandleFunc(fmt.Sprintf("/%s/ws-client", session), cp.wsClient())
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

		fmt.Println(req.IdentityId)
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

func (cp *ClientPanel) wsClient() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		go cp.wsRead(c)
		go cp.wsWrite(c)

		// #TODO: handle exit
		for {
		}
	}
}

func (cp *ClientPanel) wsRead(c *websocket.Conn) {
	fmt.Println("CLIENT READER STARTED")
	for {
		var msg Msg
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		case "reset-clients":
			//			s.resetAllClients()
		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

func (cp *ClientPanel) wsWrite(c *websocket.Conn) {
	/*
		statusTicker := time.NewTicker(statusSendPeriod)
		defer func() {
			statusTicker.Stop()
		}()
	*/
	fmt.Println("CLIENT WRITER STARTED")
	for {
		select {
		//		case <-done:
		//			return
		//		case <-statusTicker.C:
		//			fmt.Println("TICKER!")

		case status := <-cp.chSendServerStatus:
			msg := Msg{Type: "server-status", ServerStatus: status}
			err := c.WriteJSON(&msg)
			if err != nil {
				// #TODO: relax
				log.Fatal(err)
			}
		}
	}
}
