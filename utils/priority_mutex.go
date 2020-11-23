package utils

import (
	"sync"
)

// PriorityMutex is a special type of mutex
// that allows callers to request priority
// over other callers. This can be useful
// if there is a "hot path" in an application
// that requires lock access.
//
// WARNING: It is possible to cause lock starvation
// if not careful (i.e. only high priority callers
// ever do work).
type PriorityMutex struct {
	high []chan struct{}
	low  []chan struct{}

	m sync.Mutex
	l bool
}

// NewPriorityMutex returns a new *PriorityMutex.
func NewPriorityMutex() *PriorityMutex {
	return &PriorityMutex{
		high: []chan struct{}{},
		low:  []chan struct{}{},
	}
}

// Lock attempts to acquire either a high or low
// priority mutex. When priority is true, a lock
// will be granted before other low priority callers.
func (m *PriorityMutex) Lock(priority bool) {
	m.m.Lock()

	if !m.l {
		m.l = true
		m.m.Unlock()
		return
	}

	c := make(chan struct{})
	if priority {
		m.high = append(m.high, c)
	} else {
		m.low = append(m.low, c)
	}

	m.m.Unlock()
	<-c
}

// Unlock selects the next highest priority lock
// to grant. If there are no locks to grant, it
// sets the value of m.l to false.
func (m *PriorityMutex) Unlock() {
	m.m.Lock()
	defer m.m.Unlock()

	if len(m.high) > 0 {
		c := m.high[0]
		m.high = m.high[1:]
		close(c)
		return
	}

	if len(m.low) > 0 {
		c := m.low[0]
		m.low = m.low[1:]
		close(c)
		return
	}

	// We only set m.l to false when there are
	// no items to unlock because it could cause
	// lock contention for the next lock to fetch it.
	m.l = false
}
