package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Identity struct {
	DisplayName string `json:"displayName"`
}

func makeIdentity(workDir string, displayName string) error {
	uuid := uuid.New()

	identity := Identity{
		DisplayName: displayName,
	}

	data, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	filepath := filepath.Clean(fmt.Sprintf("%sidentity-%s.json", workDir, uuid.String()))

	return os.WriteFile(filepath, data, 0644)
}

// uuid can be ""
func getIdentity(workDir string, uuid string) (*Identity, error) {
	identityFilePath := ""
	if uuid == "" {
		filematches, err := filepath.Glob(workDir + "identity-*.json")
		if err != nil {
			return nil, err
		}

		if len(filematches) == 0 {
			return nil, fmt.Errorf("No identity file found - ensure you've run this program with make-identity first")
		}

		if len(filematches) > 1 {
			return nil, fmt.Errorf("Multiple identity files found - please specify")
		}

		identityFilePath = filematches[0]
	} else {
		identityFilePath = filepath.Clean(fmt.Sprintf("%sidentity-%s.json", workDir, uuid))
	}

	data, err := os.ReadFile(identityFilePath)
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
