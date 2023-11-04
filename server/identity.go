package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer/config"
	"github.com/creativenucleus/bytejammer/crypto"
	"github.com/google/uuid"
)

type JammerIdentity struct {
	Uuid         uuid.UUID            `json:"uuid"`
	DisplayName  string               `json:"displayName"`
	CryptoPublic *crypto.CryptoPublic `json:"cryptoPublic"`
	CreatedAt    time.Time            `json:"createdAt"`
}

type Identities struct {
	identities map[string]JammerIdentity
}

// Loads in the known identities from disk...
func NewIdentities() (*Identities, error) {
	i := Identities{
		identities: make(map[string]JammerIdentity),
	}

	filematches, err := filepath.Glob(fmt.Sprintf("%sserver-data/identity/identity-*.json", config.WORK_DIR))
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
		i.identities[key] = *identity
	}

	return &i, nil
}

func (i *Identities) getIdentityById(id string) *JammerIdentity {
	identity, ok := i.identities[id]
	if !ok {
		return nil
	}

	return &identity
}

func (i *Identities) addIdentity(identity JammerIdentity) error {
	data, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	filepath := filepath.Clean(fmt.Sprintf("%sserver-data/identity/identity-%s.json", config.WORK_DIR, identity.Uuid.String()))

	return os.WriteFile(filepath, data, 0644)
}

func readIdentityFile(filepath string) (*JammerIdentity, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var identity JammerIdentity
	err = json.Unmarshal(data, &identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}
