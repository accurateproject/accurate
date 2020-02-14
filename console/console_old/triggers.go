package console

import (
	"github.com/accurateproject/accurate/apier/v1"
	"github.com/accurateproject/accurate/engine"
)

func init() {
	c := &CmdGetTriggers{
		name:      "triggers",
		rpcMethod: "ApiV1.GetActionTriggers",
		rpcParams: &v1.AttrGetActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetActionTriggers
	*CommandExecuter
}

func (self *CmdGetTriggers) Name() string {
	return self.name
}

func (self *CmdGetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetTriggers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdGetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetTriggers) RpcResult() interface{} {
	atr := engine.ActionTriggers{}
	return &atr
}
