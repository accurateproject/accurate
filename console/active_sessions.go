package console

import (
	"github.com/accurateproject/accurate/sessionmanager"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdActiveSessions{
		name:      "active_sessions",
		rpcMethod: "SessionManagerV1.ActiveSessions",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdActiveSessions struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrGetSMASessions
	*CommandExecuter
}

func (self *CmdActiveSessions) Name() string {
	return self.name
}

func (self *CmdActiveSessions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdActiveSessions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrGetSMASessions{}
	}
	return self.rpcParams
}

func (self *CmdActiveSessions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdActiveSessions) RpcResult() interface{} {
	var sessions *[]*sessionmanager.ActiveSession
	return &sessions
}
