package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type NusanLauncher struct {
	ch      *chan string
	wsConn  *websocket.Conn
	wsMutex sync.Mutex
}

func NusanLauncherConnect(port int) (*NusanLauncher, error) {
	log.Printf("-> Starting socket for NUSAN LAUNCHER on port: %d", port)

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	ch := make(chan string)
	nl := NusanLauncher{
		ch: &ch,
	}
	http.HandleFunc("/bytejammer", wsNusan(nl))

	// #TODO: This is a bit iffy - nl may be available before the connection can be used
	go func() {
		if err := webServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	return &nl, nil
}

func wsNusan(nl NusanLauncher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		nl.wsConn, err = wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("ERR upgrade:", err)
			return
		}
		// #TODO: Not great!
		defer nl.wsConn.Close()

		go nl.nusanWsOperatorRead()
		go nl.nusanWsOperatorWrite()

		// #TODO: handle exit
		for {
		}
	}
}

// #TODO: Is this used?
func (nl *NusanLauncher) nusanWsOperatorRead() {
	for {
		/*
			var msg interface{}
			err := nl.conn.ReadJSON(&msg)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("-> NUSAN LAUNCHER: %v\n", msg)
		*/
	}

}

type NusanLauncherMsg struct {
	Data struct {
		RoomName string `json:"RoomName"`
		NickName string `json:"NickName"`
	} `json:"Data"`
}

func (nl *NusanLauncher) nusanWsOperatorWrite() {
	for {
		select {
		case msg := <-(*nl.ch):
			fmt.Printf("-> NUSAN TOSEND: %v\n", msg)
			nlMsg := NusanLauncherMsg{}
			nlMsg.Data.RoomName = "bytejammer"
			nlMsg.Data.NickName = msg
			err := nl.sendData(&nlMsg)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (nl *NusanLauncher) sendData(data interface{}) error {
	nl.wsMutex.Lock()
	defer nl.wsMutex.Unlock()
	return nl.wsConn.WriteJSON(data)
}
