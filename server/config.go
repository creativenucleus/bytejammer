package server

import (
	"github.com/google/uuid"
)

type ConfigMachines struct {
	MachineUuid uuid.UUID  `json:"machine_uuid,error"`
	UserUuid    *uuid.UUID `json:"user_uuid,omitempty"`
}

type Config struct {
	Machines []ConfigMachines `json:"machines"`
}

func getConfig(m *SessionManager) Config {
	panic("not implemented")
}
