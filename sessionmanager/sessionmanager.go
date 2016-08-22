package sessionmanager

import (
	"time"

	"github.com/accurateproject/accurate/engine"
	"github.com/cgrates/rpcclient"
)

type SessionManager interface {
	Rater() rpcclient.RpcClientConnection
	CdrSrv() rpcclient.RpcClientConnection
	DebitInterval() time.Duration
	DisconnectSession(engine.Event, string, string) error
	WarnSessionMinDuration(string, string)
	Sessions() []*Session
	Timezone() string
	Connect() error
	Shutdown() error
	//RemoveSession(string)
	//SyncSessions() error
}
