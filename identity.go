package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/creativenucleus/bytejammer/config"
	"github.com/google/uuid"
)

type Identity struct {
	DisplayName string `json:"displayName"`
}

func makeIdentity(displayName string) error {
	uuid := uuid.New()

	identity := Identity{
		DisplayName: displayName,
	}

	data, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	filepath := filepath.Clean(fmt.Sprintf("%sidentity-%s.json", config.WORK_DIR, uuid.String()))

	return os.WriteFile(filepath, data, 0644)
}

func getIdentities() (map[string]Identity, error) {
	identities := make(map[string]Identity, 0)

	filematches, err := filepath.Glob(config.WORK_DIR + "identity-*.json")
	if err != nil {
		return nil, err
	}

	for _, filematch := range filematches {
		identity, err := readIdentityFile(filematch)
		if err != nil {
			return nil, err
		}

		// #TODO: This is a bit hacky
		strlen := len(filematch)
		key := filematch[strlen-41 : strlen-5]
		identities[key] = *identity
	}

	return identities, nil
}

// uuid can be ""
func getIdentity(uuid string) (*Identity, error) {
	identityFilePath := ""
	if uuid == "" {
		filematches, err := filepath.Glob(config.WORK_DIR + "identity-*.json")
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
		identityFilePath = filepath.Clean(fmt.Sprintf("%sidentity-%s.json", config.WORK_DIR, uuid))
	}

	identity, err := readIdentityFile(identityFilePath)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

func readIdentityFile(filepath string) (*Identity, error) {
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
