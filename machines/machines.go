package machines

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/creativenucleus/bytejammer/server"
	"github.com/google/uuid"
)

type Machine struct {
	MachineName string
	Platform    string
	Tic         *Tic
	Uuid        uuid.UUID
}

var MACHINES []*Machine

func GetMachine(findUuid uuid.UUID) *Machine {
	for _, m := range MACHINES {
		if m.Uuid == findUuid {
			return m
		}
	}

	return nil
}

// Ensure Machine.shutdown is called (maybe deferred?)
func LaunchMachine(platform string, hasImport bool, hasExport bool, isServer bool) (*Machine, error) {
	m := Machine{
		Platform:    platform,
		Uuid:        uuid.New(),
		MachineName: server.GetFunName(len(MACHINES)),
	}

	var err error
	switch m.Platform {
	case "TIC-80":
		slug := fmt.Sprint(rand.Intn(10000))
		m.Tic, err = newTic(slug, hasImport, hasExport, isServer)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("Unhandled platform")
	}

	MACHINES = append(MACHINES, &m)

	return &m, nil
}

func ShutdownMachine(uuidString string) error {
	findUuid, err := uuid.Parse(uuidString)
	if err != nil {
		return err
	}

	for i, m := range MACHINES {
		if m.Uuid == findUuid {
			m.Shutdown()
			MACHINES = append(MACHINES[:i], MACHINES[i+1:]...)
			return nil
		}
	}

	return nil
}

func (m *Machine) Shutdown() {
	m.Tic.shutdown()
}
