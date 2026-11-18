//go:build windows

package mlock

import (
	"errors"
	"log"
	"os"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

type pageLocker struct {
	mu       sync.Mutex
	pages    map[uintptr]struct{}
	pins     map[*byte]*runtime.Pinner
	pageSize uintptr
	lock     func(uintptr, uintptr) error
	unlock   func(uintptr, uintptr) error
}

var lockedMemory = pageLocker{
	pages:    make(map[uintptr]struct{}),
	pins:     make(map[*byte]*runtime.Pinner),
	pageSize: uintptr(os.Getpagesize()),
	lock:     windows.VirtualLock,
	unlock:   windows.VirtualUnlock,
}

func Mlock(logger *log.Logger) error {
	return errors.New("process-wide memory locking is not supported on Windows; lock individual values instead")
}

func Mlock2(logger *log.Logger, sensitive *string) error {
	return lockedMemory.protect(sensitive)
}

func (l *pageLocker) protect(sensitive *string) error {
	if sensitive == nil {
		return errors.New("cannot lock a nil string")
	}
	if len(*sensitive) == 0 {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	data := unsafe.StringData(*sensitive)
	pin, pinned := l.pins[data]
	if !pinned {
		pin = new(runtime.Pinner)
		pin.Pin(data)
	}
	address := uintptr(unsafe.Pointer(data))
	size := uintptr(len(*sensitive))
	if err := l.lock(address, size); err != nil {
		if !pinned {
			pin.Unpin()
		}
		return err
	}
	l.pins[data] = pin
	for page, last := address-address%l.pageSize, (address+size-1)/l.pageSize*l.pageSize; page <= last; page += l.pageSize {
		l.pages[page] = struct{}{}
	}
	return nil
}

func MunlockAll(logger *log.Logger) error {
	return lockedMemory.release()
}

func (l *pageLocker) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result error
	// Windows does not count repeated locks. Unlock shared pages only once.
	for page := range l.pages {
		if err := l.unlock(page, l.pageSize); err != nil {
			result = errors.Join(result, err)
		} else {
			delete(l.pages, page)
		}
	}
	// Retain backing allocations if any page could not be unlocked.
	if len(l.pages) == 0 {
		for data, pin := range l.pins {
			pin.Unpin()
			delete(l.pins, data)
		}
	}
	return result
}
