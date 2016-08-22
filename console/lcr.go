package console

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdGetLcr{
		name:      "lcr",
		rpcMethod: "ApierV1.GetLcr",
		rpcParams: &engine.LcrRequest{Paginator: &utils.Paginator{}},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetLcr struct {
	name      string
	rpcMethod string
	rpcParams *engine.LcrRequest
	*CommandExecuter
}

func (self *CmdGetLcr) Name() string {
	return self.name
}

func (self *CmdGetLcr) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetLcr) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.LcrRequest{Paginator: &utils.Paginator{}}
	}
	return self.rpcParams
}

func (self *CmdGetLcr) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetLcr) RpcResult() interface{} {
	return &engine.LcrReply{}
}
