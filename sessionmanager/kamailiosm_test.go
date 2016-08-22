
package sessionmanager

import (
	"testing"
)

func TestKamSMInterface(t *testing.T) {
	var _ SessionManager = SessionManager(new(KamailioSessionManager))
}
