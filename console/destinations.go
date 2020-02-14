package console

import (
	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/engine"
)

func init() {
	c := &CmdGetDestination{
		name:      "destinations",
		rpcMethod: "ApiV1.GetDestinations",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDestination struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetMultiple
	*CommandExecuter
}

func (self *CmdGetDestination) Name() string {
	return self.name
}

func (self *CmdGetDestination) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDestination) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetMultiple{}
	}
	return self.rpcParams
}

func (self *CmdGetDestination) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDestination) RpcResult() interface{} {
	a := make([]*engine.Destination, 0)
	return &a
}
