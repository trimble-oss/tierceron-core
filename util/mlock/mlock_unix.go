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
