//go:build darwin || dragonfly || freebsd || linux || openbsd || solaris

package mlock

import (
	"errors"
	"log"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Mlock - provides locking hook for OS's that support mlock
func Mlock(logger *log.Logger) error {
	return unix.Mlockall(unix.MCL_CURRENT | unix.MCL_FUTURE)
}

func Mlock2(logger *log.Logger, sensitive *string) error {
	return lockString(sensitive, unix.Mlock)
}

// Mlock2Move locks sensitive in place. No relocation occurs on Unix, so no
// unlocked source copy exists and nothing is erased.
func Mlock2Move(logger *log.Logger, sensitive *string) error {
	return Mlock2(logger, sensitive)
}

// Mlock2Release is a no-op: initialization locks all current and future
// mappings in place, and value backing may be shared or literal storage that
// must not be erased or unlocked.
func Mlock2Release(logger *log.Logger, sensitive *string) error {
	return nil
}

// LockedAlloc returns a zeroed buffer locked in place, so sensitive data can
// be received directly into protected storage.
func LockedAlloc(size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("locked allocation size must be positive")
	}
	buf := make([]byte, size)
	if err := unix.Mlock(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// LockedFree erases buf. The pages stay locked; process-wide initialization
// locks all mappings, so unlocking here could expose unrelated heap data.
func LockedFree(buf []byte) error {
	clear(buf)
	return nil
}

func lockString(sensitive *string, lock func([]byte) error) error {
	if sensitive == nil {
		return errors.New("cannot lock a nil string")
	}
	if len(*sensitive) == 0 {
		return nil
	}
	// Mlock only reads the address and length; it never mutates string data.
	data := unsafe.Slice(unsafe.StringData(*sensitive), len(*sensitive))
	err := lock(data)
	runtime.KeepAlive(sensitive)
	return err
}

func MunlockAll(logger *log.Logger) error {
	return unix.Munlockall()
}
