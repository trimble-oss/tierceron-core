//go:build nacl || netbsd || plan9

package mlock

import (
	"errors"
	"log"
)

// Mlock - provides locking hook for OS's that don't support mlock
func Mlock(logger *log.Logger) error {
	return errors.New("memory locking is not supported on this platform")
}

func Mlock2(logger *log.Logger, sensitive *string) error {
	return errors.New("memory locking is not supported on this platform")
}

// Mlock2Move fails without erasing; no locked copy of the value is made.
func Mlock2Move(logger *log.Logger, sensitive *string) error {
	return Mlock2(logger, sensitive)
}

// Mlock2Release is a no-op; no memory is ever locked on this platform.
func Mlock2Release(logger *log.Logger, sensitive *string) error {
	return nil
}

// LockedAlloc is unavailable; memory cannot be locked on this platform.
func LockedAlloc(size int) ([]byte, error) {
	return nil, errors.New("memory locking is not supported on this platform")
}

// LockedFree erases buf for parity with locking platforms.
func LockedFree(buf []byte) error {
	clear(buf)
	return nil
}

func MunlockAll(logger *log.Logger) error {
	return errors.New("memory unlocking is not supported on this platform")
}
