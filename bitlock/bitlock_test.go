package bitlock

import (
	"fmt"
	"os"
	"testing"
	"time"
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
	b := InitBitMask(len(*flows))

	b.Lock((*flows)["flow1"] ^ (*flows)["flow2"] ^ (*flows)["flow3"])
	fmt.Fprintln(os.Stderr, "Locking flow1, flow2, flow3")
	b.Unlock((*flows)["flow1"] ^ (*flows)["flow2"] ^ (*flows)["flow3"])

	go func() {
		b.Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Fprintln(os.Stderr, "Locking flow1, flow6")
		b.Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()

	go func() {
		b.Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Fprintln(os.Stderr, "Locking flow4, flow5")
		b.Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()

	go func() {
		b.Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Fprintln(os.Stderr, "Locking flow1, flow6")
		b.Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		b.Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Fprintln(os.Stderr, "Locking flow1, flow6")
		b.Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		b.Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Fprintln(os.Stderr, "Locking flow1, flow6")
		b.Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		b.Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Fprintln(os.Stderr, "Locking flow4, flow5")
		b.Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		b.Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Fprintln(os.Stderr, "Locking flow4, flow5")
		b.Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()

	go func() {
		b.Lock((*flows)["flow1"] ^ (*flows)["flow6"])
		fmt.Fprintln(os.Stderr, "Locking flow1, flow6")
		b.Unlock((*flows)["flow1"] ^ (*flows)["flow6"])
	}()
	go func() {
		b.Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Fprintln(os.Stderr, "Locking flow4, flow5")
		b.Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		b.Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Fprintln(os.Stderr, "Locking flow4, flow5")
		b.Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		b.Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Fprintln(os.Stderr, "Locking flow4, flow5")
		b.Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()
	go func() {
		b.Lock((*flows)["flow4"] ^ (*flows)["flow5"])
		fmt.Fprintln(os.Stderr, "Locking flow4, flow5")
		b.Unlock((*flows)["flow4"] ^ (*flows)["flow5"])
	}()

	time.Sleep(20 * time.Second)
}
