//go:build windows

package mlock

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const workingSetHeadroom = 16 * 1024 * 1024

var (
	kernel32                 = windows.NewLazySystemDLL("kernel32.dll")
	getProcessWorkingSetSize = kernel32.NewProc("GetProcessWorkingSetSize")
	setProcessWorkingSetSize = kernel32.NewProc("SetProcessWorkingSetSize")
)

type allocation struct {
	size   uintptr
	locked bool
	// length is the protected value's byte count for content deduplication;
	// it is zero for general-purpose buffers, which are never deduplicated.
	length uintptr
	// refs counts protected owners of a value allocation. Deduplication adds
	// owners; Mlock2Release removes them and frees at zero.
	refs int
}

type memoryLocker struct {
	mu          sync.Mutex
	allocations map[uintptr]allocation
	pageSize    uintptr
	allocated   uintptr
	ensureQuota func(uintptr) error
	alloc       func(uintptr) (uintptr, error)
	free        func(uintptr) error
	lock        func(uintptr, uintptr) error
	unlock      func(uintptr, uintptr) error
}

var lockedMemory = memoryLocker{
	allocations: make(map[uintptr]allocation),
	pageSize:    uintptr(os.Getpagesize()),
	ensureQuota: ensureWorkingSet,
	alloc: func(size uintptr) (uintptr, error) {
		return windows.VirtualAlloc(0, size, windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
	},
	free: func(address uintptr) error {
		return windows.VirtualFree(address, 0, windows.MEM_RELEASE)
	},
	lock:   windows.VirtualLock,
	unlock: windows.VirtualUnlock,
}

func ensureWorkingSet(needed uintptr) error {
	if needed > ^uintptr(0)-workingSetHeadroom {
		return errors.New("locked memory working set size overflows")
	}
	required := needed + workingSetHeadroom
	handle := windows.CurrentProcess()
	var minimum, maximum uintptr
	result, _, err := getProcessWorkingSetSize.Call(uintptr(handle), uintptr(unsafe.Pointer(&minimum)), uintptr(unsafe.Pointer(&maximum)))
	if result == 0 {
		return fmt.Errorf("GetProcessWorkingSetSize: %w", err)
	}
	if minimum >= required {
		return nil
	}
	if maximum < required {
		maximum = required
	}
	result, _, err = setProcessWorkingSetSize.Call(uintptr(handle), required, maximum)
	if result == 0 {
		return fmt.Errorf("SetProcessWorkingSetSize: %w", err)
	}
	return nil
}

func Mlock(logger *log.Logger) error {
	return errors.New("process-wide memory locking is not supported on Windows; lock individual values instead")
}

func Mlock2(logger *log.Logger, sensitive *string) error {
	return lockedMemory.protect(sensitive, false)
}

// Mlock2Move protects sensitive and erases the unlocked source bytes after
// relocation. The caller must exclusively own the source backing memory; it
// must not be shared, a substring of shared data, or a string literal.
func Mlock2Move(logger *log.Logger, sensitive *string) error {
	return lockedMemory.protect(sensitive, true)
}

func (l *memoryLocker) protect(sensitive *string, eraseSource bool) error {
	if sensitive == nil {
		return errors.New("cannot lock a nil string")
	}
	if len(*sensitive) == 0 {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	address := uintptr(unsafe.Pointer(unsafe.StringData(*sensitive)))
	length := uintptr(len(*sensitive))
	for base, block := range l.allocations {
		if block.locked && address >= base && address-base < block.size &&
			length <= block.size-(address-base) {
			// Already backed by locked memory; there is no unlocked source.
			return nil
		}
	}
	// Reuse an existing locked value with identical content, so repeated
	// protection of unchanged values does not grow locked memory.
	for base, block := range l.allocations {
		if !block.locked || block.length != length {
			continue
		}
		existing := unsafe.Slice((*byte)(unsafe.Pointer(base)), int(length))
		if subtle.ConstantTimeCompare(existing, unsafe.Slice(unsafe.StringData(*sensitive), int(length))) == 1 {
			source := *sensitive
			*sensitive = unsafe.String(&existing[0], int(length))
			if eraseSource {
				clear(unsafe.Slice(unsafe.StringData(source), len(source)))
			}
			block.refs++
			l.allocations[base] = block
			return nil
		}
	}
	base, _, err := l.acquireLocked(length)
	if err != nil {
		return err
	}
	// base addresses VirtualAlloc memory outside the Go heap; the GC never
	// moves or frees it, so uintptr round trips are valid despite vet unsafeptr.
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(base)), int(length))
	source := *sensitive
	copy(buffer, source)
	*sensitive = unsafe.String(unsafe.SliceData(buffer), len(buffer))
	if eraseSource {
		clear(unsafe.Slice(unsafe.StringData(source), len(source)))
	}
	record := l.allocations[base]
	record.length = length
	record.refs = 1
	l.allocations[base] = record
	return nil
}

// Mlock2Release releases one protected owner of sensitive and blanks it. The
// allocation is erased, unlocked, and freed once its last owner releases it.
// Values the locker does not own, such as never-protected strings, are left
// untouched. Substring views of a protected value cannot be released.
func Mlock2Release(logger *log.Logger, sensitive *string) error {
	return lockedMemory.releaseValue(sensitive)
}

func (l *memoryLocker) releaseValue(sensitive *string) error {
	if sensitive == nil || len(*sensitive) == 0 {
		return nil
	}
	address := uintptr(unsafe.Pointer(unsafe.StringData(*sensitive)))
	l.mu.Lock()
	defer l.mu.Unlock()
	block, owned := l.allocations[address]
	if !owned || !block.locked || block.refs == 0 {
		// Not a protected value owned by this locker; nothing to release.
		return nil
	}
	if uintptr(len(*sensitive)) != block.length {
		return errors.New("cannot release a partial view of a protected value")
	}
	block.refs--
	l.allocations[address] = block
	*sensitive = ""
	if block.refs > 0 {
		return nil
	}
	clear(unsafe.Slice((*byte)(unsafe.Pointer(address)), int(block.size)))
	if err := l.unlock(address, block.size); err != nil {
		// Erased but still locked; retained for a later retry.
		return fmt.Errorf("VirtualUnlock: %w", err)
	}
	block.locked = false
	l.allocations[address] = block
	if err := l.free(address); err != nil {
		return fmt.Errorf("VirtualFree: %w", err)
	}
	delete(l.allocations, address)
	l.allocated -= block.size
	return nil
}

// acquireLocked allocates and locks a dedicated region holding length bytes
// and records it. The caller must hold l.mu.
func (l *memoryLocker) acquireLocked(length uintptr) (uintptr, uintptr, error) {
	if length > ^uintptr(0)-(l.pageSize-1) {
		return 0, 0, errors.New("locked memory allocation size overflows")
	}
	size := (length + l.pageSize - 1) / l.pageSize * l.pageSize
	if size > uintptr(^uint(0)>>1) {
		return 0, 0, errors.New("locked memory allocation is too large")
	}
	if size > ^uintptr(0)-l.allocated {
		return 0, 0, errors.New("total locked memory size overflows")
	}
	if err := l.ensureQuota(l.allocated + size); err != nil {
		return 0, 0, err
	}
	base, err := l.alloc(size)
	if err != nil {
		return 0, 0, fmt.Errorf("VirtualAlloc (%d bytes): %w", size, err)
	}
	if err := l.lock(base, size); err != nil {
		lockErr := fmt.Errorf("VirtualLock (%d bytes): %w", size, err)
		if freeErr := l.free(base); freeErr != nil {
			l.allocations[base] = allocation{size: size}
			l.allocated += size
			return 0, 0, errors.Join(lockErr, fmt.Errorf("VirtualFree: %w", freeErr))
		}
		return 0, 0, lockErr
	}
	l.allocations[base] = allocation{size: size, locked: true}
	l.allocated += size
	return base, size, nil
}

// LockedAlloc returns a zeroed buffer backed by dedicated locked memory, so
// sensitive data can be received directly into protected storage. Release it
// with LockedFree, or process-wide with MunlockAll.
func LockedAlloc(size int) ([]byte, error) {
	return lockedMemory.allocate(size)
}

func (l *memoryLocker) allocate(size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("locked allocation size must be positive")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	base, _, err := l.acquireLocked(uintptr(size))
	if err != nil {
		return nil, err
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(base)), size), nil
}

// LockedFree erases and releases the locked allocation backing buf.
func LockedFree(buf []byte) error {
	return lockedMemory.releaseBuffer(buf)
}

func (l *memoryLocker) releaseBuffer(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	address := uintptr(unsafe.Pointer(unsafe.SliceData(buf)))
	l.mu.Lock()
	defer l.mu.Unlock()
	for base, block := range l.allocations {
		if address >= base && address-base < block.size {
			clear(unsafe.Slice((*byte)(unsafe.Pointer(base)), int(block.size)))
			if block.locked {
				if err := l.unlock(base, block.size); err != nil {
					// Erased but still locked; retained for a later retry.
					return fmt.Errorf("VirtualUnlock: %w", err)
				}
				block.locked = false
				l.allocations[base] = block
			}
			if err := l.free(base); err != nil {
				return fmt.Errorf("VirtualFree: %w", err)
			}
			delete(l.allocations, base)
			l.allocated -= block.size
			return nil
		}
	}
	return errors.New("buffer is not backed by locked memory")
}

func MunlockAll(logger *log.Logger) error {
	return lockedMemory.release()
}

func (l *memoryLocker) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var result error
	for base, block := range l.allocations {
		if block.locked {
			clear(unsafe.Slice((*byte)(unsafe.Pointer(base)), int(block.size)))
			if err := l.unlock(base, block.size); err != nil {
				result = errors.Join(result, fmt.Errorf("VirtualUnlock: %w", err))
				continue
			}
			block.locked = false
			l.allocations[base] = block
		}
		if err := l.free(base); err != nil {
			result = errors.Join(result, fmt.Errorf("VirtualFree: %w", err))
			continue
		}
		delete(l.allocations, base)
		l.allocated -= block.size
	}
	return result
}
