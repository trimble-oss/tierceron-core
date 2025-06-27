package bitlock

import (
	"fmt"
	"testing"
)

var flows *map[string]uint64

func generateFlows() map[string]uint64 {
	return map[string]uint64{
		"flow1": 1,
		"flow2": 2,
		"flow3": 4,
		"flow4": 8,
		"flow5": 16,
		"flow6": 32,
		"flow7": 64,
	}
}

func TestBitLock(t *testing.T) {
	flowGens := generateFlows()
	flows = &flowGens
	InitBitMask(len(*flows))

	Lock((*flows)["flow1"] ^ (*flows)["flow2"] ^ (*flows)["flow3"])
	fmt.Println("Locking flow1, flow2, flow3")
	Unlock((*flows)["flow1"] ^ (*flows)["flow2"] ^ (*flows)["flow3"])

	go func() {
		Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Println("Locking flow1, flow6")
		Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()

	go func() {
		Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Println("Locking flow4, flow5")
		Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()

	go func() {
		Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Println("Locking flow1, flow6")
		Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Println("Locking flow1, flow6")
		Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Println("Locking flow1, flow6")
		Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Println("Locking flow4, flow5")
		Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Println("Locking flow4, flow5")
		Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()

	go func() {
		Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Println("Locking flow1, flow6")
		Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Println("Locking flow4, flow5")
		Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Println("Locking flow4, flow5")
		Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Println("Locking flow4, flow5")
		Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Println("Locking flow4, flow5")
		Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()

	wait := make(chan bool)
	wait <- true
}
