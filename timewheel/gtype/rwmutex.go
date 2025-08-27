package gtype

import "sync"

type RWMutex struct {
	sync.RWMutex
	safe bool
}

func NewRWMutex(unsafe ...bool) *RWMutex {
	mu := new(RWMutex)
	if len(unsafe) > 0 {
		mu.safe = !unsafe[0]
	} else {
		mu.safe = true
	}
	return mu
}

func (mu *RWMutex) IsSafe() bool {
	return mu.safe
}

func (mu *RWMutex) Lock(force ...bool) {
	if mu.safe || (len(force) > 0 && force[0]) {
		mu.RWMutex.Lock()
	}
}

func (mu *RWMutex) Unlock(force ...bool) {
	if mu.safe || (len(force) > 0 && force[0]) {
		mu.RWMutex.Unlock()
	}
}

func (mu *RWMutex) RLock(force ...bool) {
	if mu.safe || (len(force) > 0 && force[0]) {
		mu.RWMutex.RLock()
	}
}

func (mu *RWMutex) RUnlock(force ...bool) {
	if mu.safe || (len(force) > 0 && force[0]) {
		mu.RWMutex.RUnlock()
	}
}
