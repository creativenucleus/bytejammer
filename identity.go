package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Identity struct {
	DisplayName string `json:"displayName"`
}

func makeIdentity(workDir string, displayName string) error {
	identity := Identity{
		DisplayName: displayName,
	}

	data, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	filepath := filepath.Clean(workDir + "/identity.json")
	return os.WriteFile(filepath, data, 0644)
}

func getIdentity(workDir string) (*Identity, error) {
	filepath := filepath.Clean(workDir + "/identity.json")
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var identity Identity
	err = json.Unmarshal(data, &identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}
