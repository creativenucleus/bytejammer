package server

import (
	"errors"

	"github.com/creativenucleus/bytejammer/machines"
	"github.com/google/uuid"
)

type SessionManager struct {
	machines       map[uuid.UUID]*machines.Machine
	conns          map[uuid.UUID]*SessionConn
	machineConnMap map[*machines.Machine]*SessionConn

	// #TODO: make this work...
	// Is this the right level??
	//	broadcaster *NusanLauncher
}

func makeSessionManager() *SessionManager {
	return &SessionManager{
		machines:       make(map[uuid.UUID]*machines.Machine),
		conns:          make(map[uuid.UUID]*SessionConn),
		machineConnMap: make(map[*machines.Machine]*SessionConn),
	}
}

// #TODO: Mutexes

func (sm *SessionManager) registerMachine(machine *machines.Machine) {
	sm.machines[machine.Uuid] = machine
}

func (sm *SessionManager) unregisterMachine(machine *machines.Machine) {
	delete(sm.machines, machine.Uuid)
}

func (sm *SessionManager) getMachine(uuid uuid.UUID) *machines.Machine {
	machine, ok := sm.machines[uuid]
	if !ok {
		return nil
	}
	return machine
}

func (sm *SessionManager) registerConn(conn *SessionConn) {
	sm.conns[conn.connUuid] = conn
}

func (sm *SessionManager) unregisterConn(conn *SessionConn) {
	delete(sm.conns, conn.connUuid)
}

func (sm *SessionManager) getConn(connUuid uuid.UUID) *SessionConn {
	conn, ok := sm.conns[connUuid]
	if !ok {
		return nil
	}
	return conn
}

// You must register a machine and conn before linking them
func (sm *SessionManager) linkMachineToConn(machineUuid uuid.UUID, connUuid uuid.UUID) error {
	machine := sm.getMachine(machineUuid)
	if machine == nil {
		return errors.New("machine not found")
	}

	conn := sm.getConn(connUuid)
	if conn == nil {
		return errors.New("conn not found")
	}

	sm.machineConnMap[machine] = conn
	return nil
}

// You must unlink a machine and conn before destroying either
func (sm *SessionManager) unlinkMachineFromConn(machineUuid uuid.UUID, connUuid uuid.UUID) error {
	machine := sm.getMachine(machineUuid)
	if machine == nil {
		return errors.New("machine not found")
	}

	conn := sm.getConn(connUuid)
	if conn == nil {
		return errors.New("conn not found")
	}

	if sm.machineConnMap[machine] != conn {
		return errors.New("machine does not link to expected conn")
	}

	delete(sm.machineConnMap, machine)
	return nil
}

func (sm *SessionManager) getConnForMachine(machine *machines.Machine) *SessionConn {
	conn, ok := sm.machineConnMap[machine]
	if !ok {
		return nil
	}
	return conn
}

func (sm *SessionManager) getMachineForConn(conn *SessionConn) *machines.Machine {
	for machine, c := range sm.machineConnMap {
		if c == conn {
			return machine
		}
	}
	return nil
}
