//go:build darwin

package memprotectopts

import (
	"log"
	"os"

	"github.com/trimble-oss/tierceron-core/v2/util/mlock"
)

func MemProtectInit(logger *log.Logger) error {
	return mlock.Mlock(logger)
}

func SetChattr(f *os.File) error {
	return nil
}

func UnsetChattr(f *os.File) error {
	return nil
}

func MemUnprotectAll(logger *log.Logger) error {
	return mlock.MunlockAll(logger)
}

func MemProtect(logger *log.Logger, sensitive *string) error {
	return mlock.Mlock2(logger, sensitive)
}
