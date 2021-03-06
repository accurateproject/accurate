package console

import (
	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdGetDataCost{
		name:       "datacost",
		rpcMethod:  "ApiV1.GetDataCost",
		clientArgs: []string{"Direction", "Category", "Tenant", "Account", "Subject", "StartTime", "Usage"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetDataCost struct {
	name       string
	rpcMethod  string
	rpcParams  *v1.AttrGetDataCost
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetDataCost) Name() string {
	return self.name
}

func (self *CmdGetDataCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetDataCost) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetDataCost{Direction: utils.OUT}
	}
	return self.rpcParams
}

func (self *CmdGetDataCost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetDataCost) RpcResult() interface{} {
	return &engine.DataCost{}
}

func (self *CmdGetDataCost) ClientArgs() []string {
	return self.clientArgs
}
