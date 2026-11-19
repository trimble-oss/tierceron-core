//go:build windows

package mlock

import (
	"errors"
	"os"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func testMemoryLocker() *memoryLocker {
	return &memoryLocker{
		allocations: make(map[uintptr]allocation),
		pageSize:    4096,
		ensureQuota: func(uintptr) error { return nil },
		alloc: func(size uintptr) (uintptr, error) {
			return windows.VirtualAlloc(0, size, windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
		},
		free:   func(address uintptr) error { return windows.VirtualFree(address, 0, windows.MEM_RELEASE) },
		lock:   func(uintptr, uintptr) error { return nil },
		unlock: func(uintptr, uintptr) error { return nil },
	}
}

func TestMemoryLockerSharesAllocation(t *testing.T) {
	locker := testMemoryLocker()
	value := strings.Repeat("x", 32)
	if err := locker.protect(&value, false); err != nil {
		t.Fatal(err)
	}
	part := value[8:]
	if err := locker.protect(&part, false); err != nil {
		t.Fatal(err)
	}
	if part != strings.Repeat("x", 24) || len(locker.allocations) != 1 {
		t.Fatalf("substring did not reuse locked allocation: %d allocations", len(locker.allocations))
	}
	if err := locker.release(); err != nil {
		t.Fatal(err)
	}
	if len(locker.allocations) != 0 || locker.allocated != 0 {
		t.Fatal("release retained allocations")
	}
}

func TestMemoryLockerLockFailure(t *testing.T) {
	locker := testMemoryLocker()
	want := errors.New("lock denied")
	locker.lock = func(uintptr, uintptr) error { return want }
	value := strings.Repeat("x", 32)
	if err := locker.protect(&value, false); !errors.Is(err, want) {
		t.Fatalf("got %v; want %v", err, want)
	}
	if len(locker.allocations) != 0 || value != strings.Repeat("x", 32) {
		t.Fatal("failed lock retained memory or changed the value")
	}
}

func TestMemoryLockerQuotaFailureDoesNotAllocate(t *testing.T) {
	locker := testMemoryLocker()
	want := errors.New("quota denied")
	locker.ensureQuota = func(uintptr) error { return want }
	locker.alloc = func(uintptr) (uintptr, error) {
		t.Fatal("allocated memory after quota failure")
		return 0, nil
	}
	value := "sensitive"
	if err := locker.protect(&value, false); !errors.Is(err, want) {
		t.Fatalf("got %v; want %v", err, want)
	}
	if value != "sensitive" || len(locker.allocations) != 0 {
		t.Fatal("quota failure changed the value or retained an allocation")
	}
}

func TestMemoryLockerMoveErasesSource(t *testing.T) {
	locker := testMemoryLocker()
	t.Cleanup(func() {
		if err := locker.release(); err != nil {
			t.Error(err)
		}
	})
	source := []byte("sensitive-material")
	value := unsafe.String(unsafe.SliceData(source), len(source))
	if err := locker.protect(&value, true); err != nil {
		t.Fatal(err)
	}
	if value != "sensitive-material" {
		t.Fatal("moved value was corrupted")
	}
	for _, b := range source {
		if b != 0 {
			t.Fatal("move did not erase the unlocked source")
		}
	}
	// A value already backed by locked memory has no source to erase.
	if err := locker.protect(&value, true); err != nil {
		t.Fatal(err)
	}
	if value != "sensitive-material" {
		t.Fatal("re-protecting a locked value erased it")
	}
}

func TestMemoryLockerAllocateAndRelease(t *testing.T) {
	locker := testMemoryLocker()
	buf, err := locker.allocate(100)
	if err != nil {
		t.Fatal(err)
	}
	if len(buf) != 100 || len(locker.allocations) != 1 {
		t.Fatalf("allocate: got %d bytes, %d allocations", len(buf), len(locker.allocations))
	}
	copy(buf, "sensitive")
	// A string viewing the buffer is recognized as already locked.
	view := unsafe.String(&buf[0], 9)
	if err := locker.protect(&view, false); err != nil {
		t.Fatal(err)
	}
	if len(locker.allocations) != 1 {
		t.Fatal("protecting a locked view allocated again")
	}
	unlocked := false
	locker.unlock = func(address, size uintptr) error {
		unlocked = true
		for _, b := range unsafe.Slice((*byte)(unsafe.Pointer(address)), int(size)) {
			if b != 0 {
				t.Error("allocation was not erased before unlocking")
				break
			}
		}
		return nil
	}
	if err := locker.releaseBuffer(buf[:50]); err != nil {
		t.Fatal(err)
	}
	if !unlocked || len(locker.allocations) != 0 || locker.allocated != 0 {
		t.Fatal("releaseBuffer did not unlock and free the allocation")
	}
	if err := locker.releaseBuffer(buf); err == nil {
		t.Fatal("releasing an untracked buffer must fail")
	}
}

func TestMemoryLockerDeduplicatesAndReleases(t *testing.T) {
	locker := testMemoryLocker()
	first := string([]byte("shared-secret"))
	second := string([]byte("shared-secret"))
	if err := locker.protect(&first, false); err != nil {
		t.Fatal(err)
	}
	if err := locker.protect(&second, true); err != nil {
		t.Fatal(err)
	}
	if len(locker.allocations) != 1 {
		t.Fatalf("equal values were not deduplicated: %d allocations", len(locker.allocations))
	}
	if unsafe.StringData(first) != unsafe.StringData(second) {
		t.Fatal("deduplicated values do not share backing")
	}
	if err := locker.releaseValue(&second); err != nil {
		t.Fatal(err)
	}
	if second != "" || len(locker.allocations) != 1 || first != "shared-secret" {
		t.Fatal("releasing one owner freed a value with remaining owners")
	}
	if err := locker.releaseValue(&first); err != nil {
		t.Fatal(err)
	}
	if first != "" || len(locker.allocations) != 0 || locker.allocated != 0 {
		t.Fatal("final release did not free the allocation")
	}
	plain := "unprotected"
	if err := locker.releaseValue(&plain); err != nil || plain != "unprotected" {
		t.Fatal("released a value the locker does not own")
	}
}

func TestMemoryLockerUnlockFailureRetainsAllocation(t *testing.T) {
	locker := testMemoryLocker()
	value := strings.Repeat("x", 32)
	if err := locker.protect(&value, false); err != nil {
		t.Fatal(err)
	}
	want := errors.New("unlock denied")
	locker.unlock = func(uintptr, uintptr) error { return want }
	if err := locker.release(); !errors.Is(err, want) {
		t.Fatalf("got %v; want %v", err, want)
	}
	if len(locker.allocations) != 1 || value != strings.Repeat("\x00", 32) {
		t.Fatal("failed unlock released backing memory or did not erase it")
	}
	locker.unlock = func(uintptr, uintptr) error { return nil }
	if err := locker.release(); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryLockerErasesBeforeUnlock(t *testing.T) {
	locker := testMemoryLocker()
	value := "sensitive"
	if err := locker.protect(&value, false); err != nil {
		t.Fatal(err)
	}
	locker.unlock = func(address, size uintptr) error {
		data := unsafe.Slice((*byte)(unsafe.Pointer(address)), int(size))
		for _, b := range data {
			if b != 0 {
				t.Error("allocation was not erased before unlocking")
				break
			}
		}
		return nil
	}
	if err := locker.release(); err != nil {
		t.Fatal(err)
	}
}

func TestWindowsMemoryIsActuallyLocked(t *testing.T) {
	locker := testMemoryLocker()
	locker.pageSize = uintptr(os.Getpagesize())
	locker.ensureQuota = ensureWorkingSet
	locker.lock = windows.VirtualLock
	locker.unlock = windows.VirtualUnlock
	t.Cleanup(func() {
		if err := locker.release(); err != nil {
			t.Error(err)
		}
	})
	value := strings.Repeat("secret", 8)
	original := unsafe.StringData(value)
	if err := locker.protect(&value, false); err != nil {
		t.Fatal(err)
	}
	if unsafe.StringData(value) == original || value != strings.Repeat("secret", 8) {
		t.Fatal("value was not copied into dedicated memory")
	}
	info := struct {
		address uintptr
		flags   uintptr
	}{address: uintptr(unsafe.Pointer(unsafe.StringData(value)))}
	if err := windows.QueryWorkingSetEx(windows.CurrentProcess(), uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		t.Fatal(err)
	}
	if info.flags&1 == 0 || info.flags&(1<<22) == 0 {
		t.Fatalf("allocated page is not locked: flags %#x", info.flags)
	}
}
