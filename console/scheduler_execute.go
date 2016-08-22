
package console

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdExecuteScheduledActions{
		name:      "scheduler_execute",
		rpcMethod: "ApierV1.ExecuteScheduledActions",
		rpcParams: &v1.AttrsExecuteScheduledActions{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdExecuteScheduledActions struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrsExecuteScheduledActions
	*CommandExecuter
}

func (self *CmdExecuteScheduledActions) Name() string {
	return self.name
}

func (self *CmdExecuteScheduledActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExecuteScheduledActions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrsExecuteScheduledActions{}
	}
	return self.rpcParams
}

func (self *CmdExecuteScheduledActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdExecuteScheduledActions) RpcResult() interface{} {
	var s string
	return &s
}
