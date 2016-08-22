package sessionmanager

import (
	"errors"
	"sync"

	"github.com/accurateproject/accurate/utils"
	"github.com/cenkalti/rpc2"
)

const CGR_CONNUUID = "cgr_connid"

var ErrConnectionNotFound = errors.New("CONNECTION_NOT_FOUND")

// Attempts to get the connId previously set in the client state container
func getClientConnId(clnt *rpc2.Client) string {
	if clnt == nil {
		return ""
	}
	uuid, hasIt := clnt.State.Get(CGR_CONNUUID)
	if !hasIt {
		return ""
	}
	return uuid.(string)
}

func NewSMGExternalConnections() *SMGExternalConnections {
	return &SMGExternalConnections{conns: make(map[string]*rpc2.Client), connMux: new(sync.Mutex)}
}

type SMGExternalConnections struct {
	conns   map[string]*rpc2.Client
	connMux *sync.Mutex
}

// Index the client connection so we can use it to communicate back
func (self *SMGExternalConnections) OnClientConnect(clnt *rpc2.Client) {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	connId := utils.GenUUID()
	clnt.State.Set(CGR_CONNUUID, connId) // Set unique id for the connection so we can identify it later in requests
	self.conns[connId] = clnt
}

// Unindex the client connection so we can use it to communicate back
func (self *SMGExternalConnections) OnClientDisconnect(clnt *rpc2.Client) {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	if connId := getClientConnId(clnt); connId != "" {
		delete(self.conns, connId)
	}
}

func (self *SMGExternalConnections) GetConnection(connId string) *rpc2.Client {
	self.connMux.Lock()
	defer self.connMux.Unlock()
	return self.conns[connId]
}
