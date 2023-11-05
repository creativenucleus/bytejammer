package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// This won't be very robust
// Needs an index.json, referencing the rest of the files

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func NewPlaylistFromZip(zipFilename string) (*Playlist, error) {
	data, err := os.ReadFile(zipFilename)
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	// Read all the files from zip archive
	files := make(map[string][]byte)
	for _, zipFile := range zipReader.File {
		files[zipFile.Name], err = readZipFile(zipFile)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	var playlist *Playlist
	indexData, ok := files[`index.json`]
	if !ok {
		fmt.Println("No index.json found, so adding zipfile contents without metadata")

		playlist = NewPlaylist()
		for location, codeData := range files {
			playlist.items = append(playlist.items, PlaylistItem{location: location, code: codeData})
		}
	} else {
		playlist, err = NewPlaylistFromJSON(indexData)
		if err != nil {
			return nil, err
		}

		for key, item := range playlist.items {
			codeData, ok := files[item.location]
			if !ok {
				return nil, fmt.Errorf("File not found (%s)", item.location)
			}

			playlist.items[key].code = codeData
		}
	}

	return playlist, nil
}
