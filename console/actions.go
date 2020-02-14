package console

import (
	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/engine"
)

func init() {
	c := &CmdGetActions{
		name:      "actions",
		rpcMethod: "ApiV1.GetActions",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetActions struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetMultiple
	*CommandExecuter
}

func (self *CmdGetActions) Name() string {
	return self.name
}

func (self *CmdGetActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetActions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetMultiple{}
	}
	return self.rpcParams
}

func (self *CmdGetActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetActions) RpcResult() interface{} {
	a := make(map[string]engine.Actions, 0)
	return &a
}
