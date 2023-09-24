package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

const (
	ORDER_ITERATE = iota
	ORDER_RANDOM
)

type PlaylistItem struct {
	location    string
	author      string
	description string
	code        []byte
}

type Playlist struct {
	order    int
	items    []PlaylistItem
	previous int
}

func NewPlaylist() *Playlist {
	return &Playlist{
		order:    ORDER_ITERATE,
		items:    []PlaylistItem{},
		previous: -1,
	}
}

func (p *Playlist) getNext() (*PlaylistItem, error) {
	if len(p.items) == 0 {
		return nil, errors.New("Playlist is empty")
	}

	var iItemToPlay int = 0
	if p.order == ORDER_ITERATE {
		iItemToPlay = (p.previous + 1) % len(p.items)
	} else { // Random
		// #TODO: ensure no repeats
		iItemToPlay = rand.Intn(len(p.items))
	}

	item := p.items[iItemToPlay]
	p.previous = iItemToPlay

	if item.code != nil {
		//		fmt.Printf("Cached: %s\n", item.location)
		return &item, nil
	}

	//	fmt.Printf("Loading: %s\n", item.location)

	respLua, err := http.Get(item.location)
	if err != nil {
		return nil, err
	}
	defer respLua.Body.Close()

	if respLua.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("ERR write: Status Code = %d", respLua.StatusCode))
	}

	data, err := io.ReadAll(respLua.Body)
	if err != nil {
		return nil, err
	}

	p.items[iItemToPlay].code = data
	return &p.items[iItemToPlay], nil
}
