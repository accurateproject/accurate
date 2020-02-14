package console

import "github.com/accurateproject/accurate/api/v1"

func init() {
	c := &CmdGetScheduledActions{
		name:      "scheduler_queue",
		rpcMethod: "ApiV1.GetScheduledActions",
		rpcParams: &v1.AttrsGetScheduledActions{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetScheduledActions struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrsGetScheduledActions
	*CommandExecuter
}

func (self *CmdGetScheduledActions) Name() string {
	return self.name
}

func (self *CmdGetScheduledActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetScheduledActions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrsGetScheduledActions{}
	}
	return self.rpcParams
}

func (self *CmdGetScheduledActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetScheduledActions) RpcResult() interface{} {
	s := make([]*v1.ScheduledActions, 0)
	return &s
}
