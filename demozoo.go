package main

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

func doDemozooAudit() error {
	playlist, err := NewPlaylistLCDZ()
	if err != nil {
		return err
	}
	playlist.order = ORDER_ITERATE

	authorCount := make(map[string]int)
	for _, item := range playlist.items {
		if _, ok := authorCount[item.author]; !ok {
			authorCount[item.author] = 1
		} else {
			authorCount[item.author]++
		}
	}

	authors := []string{}
	for author := range authorCount {
		authors = append(authors, author)
	}

	slices.Sort(authors)

	fmt.Printf("*** Total TIC %d items by %d authors\n", len(playlist.items), len(authors))
	fmt.Printf("*** By Author:\n")
	for _, author := range authors {
		fmt.Printf("%s: %d\n", author, authorCount[author])
	}

	// circuit break
	iMax := min(len(playlist.items), 1000)
	foundItems := []PlaylistItem{}
	for i := range iMax {
		playlistItem, err := playlist.getNext()
		if err != nil {
			return err
		}

		fmt.Printf("Examining %d %s\n", i, playlistItem.location)
		if strings.Contains(string(playlistItem.code), "fft(") || strings.Contains(string(playlistItem.code), "ffts(") {
			foundItems = append(foundItems, *playlistItem)
			fmt.Println("Found!")
		}

		if i >= iMax {
			fmt.Println("Circuit break")
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("*** FFT: %d\n", len(foundItems))
	for _, playlistItem := range foundItems {
		fmt.Println(playlistItem.location)
	}

	return nil
}
