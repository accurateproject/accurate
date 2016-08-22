
package sessionmanager

import (
	"testing"
)

func TestFSSMInterface(t *testing.T) {
	var _ SessionManager = SessionManager(new(FSSessionManager))
}
