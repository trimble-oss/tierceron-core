//go:build darwin
// +build darwin

package memprotectopts

import (
	"log"
	"os"

	"github.com/trimble-oss/tierceron-core/v2/util/mlock"
	"golang.org/x/sys/unix"
)

// Not a lot of effort has been put into this Darwin implementation
// for memory protection.  Som parts may be incomplete or incorrect.
func MemProtectInit(logger *log.Logger) error {
	mlock.Mlock(logger)
	return nil
}

func SetChattr(f *os.File) error {
	return nil
}

func UnsetChattr(f *os.File) error {
	return nil
}

func MemUnprotectAll(logger *log.Logger) error {
	return unix.Munlockall()
}

func MemProtect(logger *log.Logger, sensitive *string) error {
	// TODO: is this correct?
	return mlock.Mlock(logger)
}
