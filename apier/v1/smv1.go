
package v1

import (
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
)

// Interact with SessionManager
type SessionManagerV1 struct {
	SMs []sessionmanager.SessionManager // List of session managers since we support having more than one active session manager running on one host
}

func (self *SessionManagerV1) ActiveSessionMangers(ignored string, reply *[]sessionmanager.SessionManager) error {
	if len(self.SMs) == 0 {
		return utils.ErrNotFound
	}
	*reply = self.SMs
	return nil
}

func (self *SessionManagerV1) ActiveSessions(attrs utils.AttrGetSMASessions, reply *[]*sessionmanager.ActiveSession) error {
	if attrs.SessionManagerIndex > len(self.SMs)-1 {
		return utils.ErrNotFound
	}
	for _, session := range self.SMs[attrs.SessionManagerIndex].Sessions() {
		*reply = append(*reply, session.AsActiveSessions()...)
	}
	return nil
}
