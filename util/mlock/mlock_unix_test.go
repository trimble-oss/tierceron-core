//go:build darwin || dragonfly || freebsd || linux || openbsd || solaris

package mlock

import (
	"errors"
	"testing"

	"golang.org/x/sys/unix"
)

func TestLockStringPropagatesErrors(t *testing.T) {
	for _, want := range []error{unix.ENOMEM, unix.EPERM, unix.ENOSYS} {
		t.Run(want.Error(), func(t *testing.T) {
			value := "test data"
			err := lockString(&value, func(data []byte) error {
				if string(data) != value {
					t.Fatal("locking the wrong data")
				}
				return want
			})
			if !errors.Is(err, want) {
				t.Fatalf("got %v; want %v", err, want)
			}
		})
	}
}

func TestLockStringEmptyAndNil(t *testing.T) {
	lock := func([]byte) error {
		t.Fatal("no syscall should be made for empty or nil data")
		return nil
	}
	if err := lockString(nil, lock); err == nil {
		t.Fatal("nil string must be rejected")
	}
	value := ""
	if err := lockString(&value, lock); err != nil {
		t.Fatal(err)
	}
}
