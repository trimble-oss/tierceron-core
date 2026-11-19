//go:build linux

package memprotectopts

import (
	"log"
	"os"

	"github.com/trimble-oss/tierceron-core/v2/util/mlock"
)

// MemProtectInit initializes memory protection
func MemProtectInit(logger *log.Logger) error {
	return mlock.Mlock(logger)
}

func SetChattr(f *os.File) error {
	// return chattr.SetAttr(f, chattr.FS_IMMUTABLE_FL)
	return nil
}

func UnsetChattr(f *os.File) error {
	// return chattr.UnsetAttr(f, chattr.FS_IMMUTABLE_FL)
	return nil
}

// MemUnprotectAll unprotects all memory
func MemUnprotectAll(logger *log.Logger) error {
	return mlock.MunlockAll(logger)
}

// MemProtect protects sensitive memory
func MemProtect(logger *log.Logger, sensitive *string) error {
	return mlock.Mlock2(logger, sensitive)
}

// MemProtectMove protects sensitive memory and erases any unlocked source copy
// left behind by relocation. The caller must exclusively own the value's
// backing memory; it must not be shared, a substring, or a string literal.
func MemProtectMove(logger *log.Logger, sensitive *string) error {
	return mlock.Mlock2Move(logger, sensitive)
}

// MemRelease releases one protected owner of sensitive; on this platform all
// mappings stay locked in place, so this is a no-op.
func MemRelease(logger *log.Logger, sensitive *string) error {
	return mlock.Mlock2Release(logger, sensitive)
}

// LockedAlloc returns a zeroed buffer backed by locked memory, so sensitive
// data can be received directly into protected storage. Release with LockedFree.
func LockedAlloc(size int) ([]byte, error) {
	return mlock.LockedAlloc(size)
}

// LockedFree erases buf; its pages remain locked process-wide.
func LockedFree(buf []byte) error {
	return mlock.LockedFree(buf)
}
