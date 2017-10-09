package pearl

import (
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"sync"

	"github.com/mmcloughlin/pearl/torcrypto"
)

// GenerateCircID generates a circuit ID with the given most significant bit.
func GenerateCircID(f CellFormat, msb uint32) CircID {
	b := torcrypto.Rand(4)
	x := binary.BigEndian.Uint32(b)
	x = (x >> 1) | (msb << 31)
	if f.CircIDLen() == 2 {
		x >>= 16
	}
	return CircID(x)
}

type CircuitDirectionState struct {
	digest []byte
	cipher.Stream
}

func NewCircuitDirectionState(d, k []byte) CircuitDirectionState {
	return CircuitDirectionState{
		digest: d,
		Stream: torcrypto.NewStream(k),
	}
}

type Circuit struct {
	ID       CircID
	Forward  CircuitDirectionState
	Backward CircuitDirectionState
}

// CircuitManager manages a collection of circuits.
type CircuitManager struct {
	circuits map[CircID]*Circuit

	sync.RWMutex
}

func NewCircuitManager() *CircuitManager {
	return &CircuitManager{
		circuits: make(map[CircID]*Circuit),
	}
}

func (m *CircuitManager) AddCircuit(c *Circuit) error {
	m.Lock()
	defer m.Unlock()
	_, exists := m.circuits[c.ID]
	if exists {
		return errors.New("cannot override existing circuit id")
	}
	m.circuits[c.ID] = c
	return nil
}

func (m *CircuitManager) Circuit(id CircID) (*Circuit, bool) {
	m.RLock()
	defer m.RUnlock()
	c, ok := m.circuits[id]
	return c, ok
}
