package log

import (
	"bytes"
	"io"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// FilteredWriter wraps an io.Writer and applies multiple named filters
// that can be enabled/disabled dynamically. All enabled filters must
// pass for the write to occur.
type FilteredWriter struct {
	writer  io.Writer
	filters cmap.ConcurrentMap[string, func([]byte) bool]
}

// NewFilteredWriter creates a new FilteredWriter that wraps the given writer.
func NewFilteredWriter(writer io.Writer) *FilteredWriter {
	return &FilteredWriter{
		writer:  writer,
		filters: cmap.New[func([]byte) bool](),
	}
}

// Write implements io.Writer. It applies all enabled filters before writing.
// If any filter returns false, the write is suppressed (but reported as successful).
// Special case: if the bytes contain "<nofilter>", that tag is removed and
// the remaining bytes are written without any filtering.
func (fw *FilteredWriter) Write(p []byte) (n int, err error) {
	// Check for bypass tag
	noFilterTag := []byte("<nofilter>")
	if bytes.Contains(p, noFilterTag) {
		// Remove the tag and write without filtering
		cleaned := bytes.ReplaceAll(p, noFilterTag, nil)
		_, err := fw.writer.Write(cleaned)
		if err != nil {
			return 0, err
		}
		// Return original length to satisfy io.Writer contract
		return len(p), nil
	}

	// Check all enabled filters
	passesAllFilters := true
	for item := range fw.filters.IterBuffered() {
		if filter := item.Val; filter != nil {
			if !filter(p) {
				passesAllFilters = false
				break
			}
		}
	}

	if passesAllFilters {
		return fw.writer.Write(p)
	}

	// Filtered out - pretend we wrote it to avoid errors upstream
	return len(p), nil
}

// EnableFilter adds or updates a named filter. The filter function should
// return true to allow the write, false to suppress it.
func (fw *FilteredWriter) EnableFilter(name string, filter func([]byte) bool) {
	fw.filters.Set(name, filter)
}

// DisableFilter removes a named filter.
func (fw *FilteredWriter) DisableFilter(name string) {
	fw.filters.Remove(name)
}

// ClearFilters removes all filters.
func (fw *FilteredWriter) ClearFilters() {
	fw.filters.Clear()
}

// HasFilter checks if a named filter is currently enabled.
func (fw *FilteredWriter) HasFilter(name string) bool {
	return fw.filters.Has(name)
}

// FilterCount returns the number of currently enabled filters.
func (fw *FilteredWriter) FilterCount() int {
	return fw.filters.Count()
}
