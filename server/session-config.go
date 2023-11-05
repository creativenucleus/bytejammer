package server

import "github.com/google/uuid"

type ConfigMachine struct {
	Name           string     `json:"name"`
	Uuid           uuid.UUID  `json:"uuid"`
	JammerIdentity *uuid.UUID `json:"jammer_identity,omitempty"`
}

type SessionConfig struct {
	Port     int             `json:"port"`
	Name     string          `json:"name"`
	Slug     string          `json:"slug"`
	Machines []ConfigMachine `json:"machines"`
}

// JamSessionConfig should be enough to save to disk and restart a JamSession if it crashes
func getSessionConfig(s Session) SessionConfig {
	sc := SessionConfig{
		Port: s.port,
		Name: s.name,
		Slug: s.slug,
	}

	for _, machine := range s.switchboard.machines {
		var jammerIdentity *uuid.UUID
		machineConn, ok := s.switchboard.machineConnMap[machine]
		if ok {
			if machineConn.identity != nil {
				jammerIdentity = &machineConn.identity.uuid
			}
		}

		sc.Machines = append(sc.Machines, ConfigMachine{
			Name:           machine.MachineName,
			Uuid:           machine.Uuid,
			JammerIdentity: jammerIdentity,
		})
	}

	return sc
}
