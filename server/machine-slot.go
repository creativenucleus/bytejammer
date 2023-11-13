package server

// MachineSlots are holders for machines on a server - they can also be placeholders that
// describe a machine that hasn't yet launched, so a server can be prepared with a bunch before
// people join

import (
	"fmt"
	"sync"

	"github.com/creativenucleus/bytejammer/machines"
	"github.com/google/uuid"
)

var (
	MAX_MACHINE_SLOTS = 4
)

type MachineSlot struct {
	// This is null if the machine has not been linked
	// A machine may be linked to a client, but not be open, and the client may be disconnected
	// (This is used to reattach disconnected clients when they rejoin)
	jammerIdentityUUID uuid.NullUUID

	// "" if there's no restriction
	platformReservation string

	// nil if not launched
	machine *machines.Machine
}

// machine may be nil
func (ms *MachineSlot) AssignMachine(machine *machines.Machine) {
	ms.machine = machine
}

func (ms *MachineSlot) ReserveForJammer(jammerIdentityUUID uuid.UUID) {
	ms.jammerIdentityUUID = uuid.NullUUID{
		UUID:  jammerIdentityUUID,
		Valid: true,
	}
}

func (ms *MachineSlot) UnreserveJammers() {
	ms.jammerIdentityUUID = uuid.NullUUID{
		Valid: false,
	}
}

var MACHINE_SLOTS_MUTEX sync.Mutex
var MACHINE_SLOTS []*MachineSlot

func CanCreateMachineSlot() bool {
	return len(MACHINE_SLOTS) < MAX_MACHINE_SLOTS
}

// Mutex for extra caution
func CreateMachineSlot() (*MachineSlot, error) {
	MACHINE_SLOTS_MUTEX.Lock()
	defer MACHINE_SLOTS_MUTEX.Unlock()

	if !CanCreateMachineSlot() {
		return nil, fmt.Errorf("requested to create a machine slot, but our limit is [%d]", MAX_MACHINE_SLOTS)
	}

	slot := MachineSlot{}
	MACHINE_SLOTS = append(MACHINE_SLOTS, &slot)
	return &slot, nil
}

// Rebuild the list without this item.
// Uses a Mutex so that we don't collide
func DestroyMachineSlot(ms *MachineSlot) {
	MACHINE_SLOTS_MUTEX.Lock()
	defer MACHINE_SLOTS_MUTEX.Unlock()

	newSlots := make([]*MachineSlot, 0)
	for _, slot := range MACHINE_SLOTS {
		if slot != ms {
			newSlots = append(newSlots, slot)
		}
	}
	MACHINE_SLOTS = newSlots
}

func GetFreeMachineSlot() *MachineSlot {
	for _, slot := range MACHINE_SLOTS {
		if slot.machine == nil {
			return slot
		}
	}

	return nil
}
