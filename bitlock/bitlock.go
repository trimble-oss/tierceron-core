package bitlock

import (
	"sync"
	"sync/atomic"
)

type BitLock struct {
	bitmask           *atomic.Uint64
	bitmaskLock       *sync.RWMutex
	bitmaskUpdateChan chan bool
	listening         bool
}

func InitBitMask(size int) *BitLock {
	if size <= 0 {
		return nil
	}
	b := &BitLock{
		bitmask:     new(atomic.Uint64),
		bitmaskLock: &sync.RWMutex{},
	}
	b.bitmask.Store(0)
	b.bitmaskUpdateChan = make(chan bool)
	return b
}

func (b *BitLock) Lock(flowID uint64) {
	if b == nil || b.bitmask == nil || b.bitmaskLock == nil || b.bitmaskUpdateChan == nil {
		return
	}

check:
	conflict := b.checkConflict(flowID)

	if conflict {
		b.bitmaskLock.Lock()
		b.listening = true
		b.bitmaskLock.Unlock()
		<-b.bitmaskUpdateChan
		goto check
	}

	b.bitmaskLock.Lock()
	currentVal := b.bitmask.Load()
	result := currentVal ^ flowID
	b.bitmask.Store(result)
	if b.listening {
		b.bitmaskUpdateChan <- true
		b.listening = false
	}
	b.bitmaskLock.Unlock()
}

func (b *BitLock) Unlock(flowID uint64) {
	if b == nil || b.bitmask == nil || b.bitmaskLock == nil || b.bitmaskUpdateChan == nil {
		return
	}

	b.bitmaskLock.Lock()
	currentVal := b.bitmask.Load()
	result := currentVal ^ flowID
	b.bitmask.Store(result)
	if b.listening {
		b.bitmaskUpdateChan <- true
		b.listening = false
	}
	b.bitmaskLock.Unlock()
}

func (b *BitLock) checkConflict(flowID uint64) bool {
	if b == nil || b.bitmask == nil || b.bitmaskLock == nil {
		return false
	}
	b.bitmaskLock.RLock()
	defer b.bitmaskLock.RUnlock()
	currentMask := b.bitmask.Load()
	return (currentMask & flowID) != 0
}
