// Package memprotectopts provides OS-backed page locking for sensitive strings.
//
// Callers must check every returned error before using or publishing a value.
// A failed lock must reject the affected operation; initialization failure must
// prevent startup in a memonly build. This package does not terminate the process.
//
// Unix initialization locks current and future mappings. Windows initialization
// verifies the API provider, and each value must subsequently be passed to
// MemProtect. Windows retains and pins protected backing allocations until
// MemUnprotectAll successfully unlocks their pages. Overlapping pages are
// unlocked only once, and failed unlocks retain their pins for a later retry.
//
// MemUnprotectAll is a process-wide lifecycle operation, not per-request cleanup.
// Do not call it while another operation still relies on locked values.
// Locking prevents paging; it does not encrypt, erase, or lock other copies.
package memprotectopts
