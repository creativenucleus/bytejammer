package main

import "encoding/json"

type PlaylistJSON struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Items       []struct {
		Location    string `json:"location"`
		Author      string `json:"author"`
		Description string `json:"description"`
	} `json:"items"`
}

func NewPlaylistFromJSON(bytesIn []byte) (*Playlist, error) {
	var playlistJson PlaylistJSON
	err := json.Unmarshal(bytesIn, &playlistJson)
	if err != nil {
		return nil, err
	}

	p := NewPlaylist()

	p.order = ORDER_ITERATE
	for _, item := range playlistJson.Items {
		p.items = append(p.items, PlaylistItem{
			location:    item.Location,
			author:      item.Author,
			description: item.Description,
		})
	}

	return p, nil
}
