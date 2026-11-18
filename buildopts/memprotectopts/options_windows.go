//go:build windows

package memprotectopts

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"github.com/trimble-oss/tierceron-core/v2/util/mlock"
	"golang.org/x/sys/windows"
)

const (
	MEM_COMMIT     = windows.MEM_COMMIT
	MEM_RESERVE    = windows.MEM_RESERVE
	PAGE_READWRITE = windows.PAGE_READWRITE
)

// MemProtectInit verifies the Windows API provider. Each value is locked separately.
func MemProtectInit(logger *log.Logger) error {
	kernelTrust.Lock()
	defer kernelTrust.Unlock()
	if kernelTrust.verified {
		return nil
	}
	if err := verifyKernelTrust(); err != nil {
		return err
	}
	kernelTrust.verified = true
	return nil
}

// MemUnprotectAll releases pages locked by MemProtect.
func MemUnprotectAll(logger *log.Logger) error {
	return mlock.MunlockAll(logger)
}

func SetChattr(f *os.File) error {
	return nil
}

func UnsetChattr(f *os.File) error {
	return nil
}

var kernelTrust struct {
	sync.Mutex
	verified bool
}

func verifyKernelTrust() error {
	dllPath := filepath.Join("C:\\Windows\\System32", "kernel32.dll")
	evsignedfile16, err := windows.UTF16PtrFromString(dllPath)
	if err != nil {
		return err
	}
	data := &windows.WinTrustData{
		Size:             uint32(unsafe.Sizeof(windows.WinTrustData{})),
		UIChoice:         windows.WTD_UI_NONE,
		RevocationChecks: windows.WTD_REVOKE_WHOLECHAIN,
		UnionChoice:      windows.WTD_CHOICE_FILE,
		StateAction:      windows.WTD_STATEACTION_VERIFY,
		FileOrCatalogOrBlobOrSgnrOrCert: unsafe.Pointer(&windows.WinTrustFileInfo{
			Size:     uint32(unsafe.Sizeof(windows.WinTrustFileInfo{})),
			FilePath: evsignedfile16,
		}),
	}
	verifyErr := windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, data)
	data.StateAction = windows.WTD_STATEACTION_CLOSE
	return errors.Join(verifyErr, windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, data))
}

// MemProtect protects sensitive memory
func MemProtect(logger *log.Logger, sensitive *string) error {
	if err := MemProtectInit(logger); err != nil {
		return err
	}
	return mlock.Mlock2(logger, sensitive)
}
