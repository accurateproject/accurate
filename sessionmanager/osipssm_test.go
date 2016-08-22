
package sessionmanager

import (
	"testing"
)

func TestOsipsSMInterface(t *testing.T) {
	var _ SessionManager = SessionManager(new(OsipsSessionManager))
}
