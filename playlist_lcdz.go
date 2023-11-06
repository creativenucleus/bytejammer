package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"golang.org/x/net/html"
)

func NewPlaylistLCDZ() (*Playlist, error) {
	urlSource := "https://livecode.demozoo.org/type/Byte%2520Jam.html"
	resp, err := http.Get(urlSource)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal(err)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	p := NewPlaylist()
	links := findAllLuaLinks(doc)
	for _, link := range links {
		p.items = append(p.items, PlaylistItem{location: link})
	}
	p.order = ORDER_RANDOM

	return p, nil
}

func findAllLuaLinks(n *html.Node) []string {
	var links []string
	re := regexp.MustCompile(`.lua$`)
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, element := range n.Attr {
			if element.Key == "href" {
				if re.MatchString(element.Val) {
					links = append(links, fmt.Sprintf("https://livecode.demozoo.org%s", element.Val))
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		newLinks := findAllLuaLinks(c)
		links = append(links, newLinks...)
	}

	return links
}
