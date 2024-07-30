package coreopts

import (
	"strings"
)

// Determines if running tierceron in the default local development mode
// with the default test host.
func IsLocalEndpoint(addr string) bool {
	return strings.HasPrefix(addr, "https://tierceron.test:1234")
}
