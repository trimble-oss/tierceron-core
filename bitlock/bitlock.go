package bitlock

import (
	"sync"
	"sync/atomic"
)

var bitmask *atomic.Uint64 = new(atomic.Uint64)

var bitmaskLock *sync.RWMutex

var bitmaskUpdateChan chan bool

var listening = false

func InitBitMask(size int) {
	if size <= 0 {
		return
	}
	bitmask.Store(0)
	bitmaskLock = &sync.RWMutex{}
	bitmaskUpdateChan = make(chan bool)
}

func Lock(flowID uint64) {
	if bitmask == nil || bitmaskLock == nil {
		return
	}

check:
	conflict := checkConflict(flowID)

	if conflict {
		bitmaskLock.Lock()
		listening = true
		bitmaskLock.Unlock()
		<-bitmaskUpdateChan
		goto check
	}

	bitmaskLock.Lock()
	currentVal := bitmask.Load()
	result := currentVal ^ flowID
	bitmask.Store(result)
	if listening {
		bitmaskUpdateChan <- true
		listening = false
	}
	bitmaskLock.Unlock()
}

func Unlock(flowID uint64) {
	if bitmask == nil || bitmaskLock == nil {
		return
	}

	bitmaskLock.Lock()
	currentVal := bitmask.Load()
	result := currentVal ^ flowID
	bitmask.Store(result)
	if listening {
		bitmaskUpdateChan <- true
		listening = false
	}
	bitmaskLock.Unlock()
}

func checkConflict(flowID uint64) bool {
	if bitmask == nil || bitmaskLock == nil {
		return false
	}
	bitmaskLock.RLock()
	defer bitmaskLock.RUnlock()
	currentMask := bitmask.Load()
	return (currentMask & flowID) != 0
}
