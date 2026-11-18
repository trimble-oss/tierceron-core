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

func MunlockAll(logger *log.Logger) error {
	return errors.New("memory unlocking is not supported on this platform")
}
