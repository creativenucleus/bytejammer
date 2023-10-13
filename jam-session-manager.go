package main

import (
	"errors"

	"github.com/creativenucleus/bytejammer/machines"
	"github.com/google/uuid"
)

type JamSessionManager struct {
	machines       map[uuid.UUID]*machines.Machine
	conns          map[uuid.UUID]*JamSessionConn
	machineConnMap map[*machines.Machine]*JamSessionConn

	// #TODO: make this work...
	// Is this the right level??
	//	broadcaster *NusanLauncher
}

func makeJamSessionManager() *JamSessionManager {
	return &JamSessionManager{
		machines:       make(map[uuid.UUID]*machines.Machine),
		conns:          make(map[uuid.UUID]*JamSessionConn),
		machineConnMap: make(map[*machines.Machine]*JamSessionConn),
	}
}

// #TODO: Mutexes

func (m *JamSessionManager) registerMachine(machine *machines.Machine) {
	m.machines[machine.Uuid] = machine
}

func (m *JamSessionManager) unregisterMachine(machine *machines.Machine) {
	delete(m.machines, machine.Uuid)
}

func (m *JamSessionManager) getMachine(uuid uuid.UUID) *machines.Machine {
	machine, ok := m.machines[uuid]
	if !ok {
		return nil
	}
	return machine
}

func (m *JamSessionManager) registerConn(conn *JamSessionConn) {
	m.conns[conn.connUuid] = conn
}

func (m *JamSessionManager) unregisterConn(conn *JamSessionConn) {
	delete(m.conns, conn.connUuid)
}

func (m *JamSessionManager) getConn(connUuid uuid.UUID) *JamSessionConn {
	conn, ok := m.conns[connUuid]
	if !ok {
		return nil
	}
	return conn
}

// You must register a machine and conn before linking them
func (m *JamSessionManager) linkMachineToConn(machineUuid uuid.UUID, connUuid uuid.UUID) error {
	machine := m.getMachine(machineUuid)
	if machine == nil {
		return errors.New("machine not found")
	}

	conn := m.getConn(connUuid)
	if conn == nil {
		return errors.New("conn not found")
	}

	m.machineConnMap[machine] = conn
	return nil
}

// You must unlink a machine and conn before destroying either
func (m *JamSessionManager) unlinkMachineFromConn(machineUuid uuid.UUID, connUuid uuid.UUID) error {
	machine := m.getMachine(machineUuid)
	if machine == nil {
		return errors.New("machine not found")
	}

	conn := m.getConn(connUuid)
	if conn == nil {
		return errors.New("conn not found")
	}

	if m.machineConnMap[machine] != conn {
		return errors.New("machine does not link to expected conn")
	}

	delete(m.machineConnMap, machine)
	return nil
}

func (m *JamSessionManager) getConnForMachine(machine *machines.Machine) *JamSessionConn {
	conn, ok := m.machineConnMap[machine]
	if !ok {
		return nil
	}
	return conn
}

func (m *JamSessionManager) getMachineForConn(conn *JamSessionConn) *machines.Machine {
	for machine, c := range m.machineConnMap {
		if c == conn {
			return machine
		}
	}
	return nil
}
