package machines

import (
	"errors"
	"fmt"
	"math/rand"
)

type Machine struct {
	Platform string
	Tic      *Tic

	// #TODO: Replace this with some pointer
	JammerName string
}

var MACHINES []*Machine

// Ensure Machine.shutdown is called (maybe deferred?)
func LaunchMachine(platform string, hasImport bool, hasExport bool, isServer bool) (*Machine, error) {
	m := Machine{
		Platform: platform,
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

func (m *Machine) Shutdown() {
	m.Tic.shutdown()
}
