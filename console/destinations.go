package console

import (
	"github.com/accurateproject/accurate/apier/v2"
	"github.com/accurateproject/accurate/engine"
)

func init() {
	c := &CmdGetDestination{
		name:      "destinations",
		rpcMethod: "ApierV2.GetDestinations",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDestination struct {
	name      string
	rpcMethod string
	rpcParams *v2.AttrGetDestinations
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
		self.rpcParams = &v2.AttrGetDestinations{}
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
