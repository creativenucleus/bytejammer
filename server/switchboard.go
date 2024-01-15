package server

import (
	"errors"

	"github.com/creativenucleus/bytejammer/machines"
	"github.com/google/uuid"
)

type Switchboard struct {
	machines       map[uuid.UUID]*machines.Machine
	conns          map[uuid.UUID]*SessionConn
	machineConnMap map[*machines.Machine]*SessionConn

	// #TODO: make this work...
	// Is this the right level??
	//	broadcaster *NusanLauncher
}

func makeSwitchboard() *Switchboard {
	return &Switchboard{
		machines:       make(map[uuid.UUID]*machines.Machine),
		conns:          make(map[uuid.UUID]*SessionConn),
		machineConnMap: make(map[*machines.Machine]*SessionConn),
	}
}

// #TODO: Mutexes

func (s *Switchboard) registerMachine(m *machines.Machine) {
	s.machines[m.Uuid] = m

	go func() {
		<-m.ChClosedErr
		s.unregisterMachine(m)
	}()
}

func (s *Switchboard) unregisterMachine(m *machines.Machine) {
	conn := s.getConnForMachine(m)
	if conn != nil {
		s.unlinkMachineFromConn(m.Uuid, conn.connUuid) // #TODO: error ignored (log instead?)
	}

	delete(s.machines, m.Uuid)
}

func (s *Switchboard) getMachine(uuid uuid.UUID) *machines.Machine {
	machine, ok := s.machines[uuid]
	if !ok {
		return nil
	}
	return machine
}

func (s *Switchboard) registerConn(conn *SessionConn) {
	s.conns[conn.connUuid] = conn
}

func (s *Switchboard) unregisterConn(conn *SessionConn) {
	delete(s.conns, conn.connUuid)
}

func (sm *Switchboard) getConn(connUuid uuid.UUID) *SessionConn {
	conn, ok := sm.conns[connUuid]
	if !ok {
		return nil
	}
	return conn
}

// You must register a machine and conn before linking them
func (s *Switchboard) linkMachineToConn(machineUuid uuid.UUID, connUuid uuid.UUID) error {
	machine := s.getMachine(machineUuid)
	if machine == nil {
		return errors.New("machine not found")
	}

	conn := s.getConn(connUuid)
	if conn == nil {
		return errors.New("conn not found")
	}

	s.machineConnMap[machine] = conn
	return nil
}

// You must unlink a machine and conn before destroying either
func (s *Switchboard) unlinkMachineFromConn(machineUuid uuid.UUID, connUuid uuid.UUID) error {
	machine := s.getMachine(machineUuid)
	if machine == nil {
		return errors.New("machine not found")
	}

	conn := s.getConn(connUuid)
	if conn == nil {
		return errors.New("conn not found")
	}

	if s.machineConnMap[machine] != conn {
		return errors.New("machine does not link to expected conn")
	}

	delete(s.machineConnMap, machine)
	return nil
}

func (s *Switchboard) getConnForMachine(machine *machines.Machine) *SessionConn {
	conn, ok := s.machineConnMap[machine]
	if !ok {
		return nil
	}
	return conn
}

func (s *Switchboard) getMachineForConn(conn *SessionConn) *machines.Machine {
	for machine, c := range s.machineConnMap {
		if c == conn {
			return machine
		}
	}
	return nil
}
