package console

import (
	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdAddBalance{
		name:      "balance_add",
		rpcMethod: "ApiV1.AddBalance",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAddBalance struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetBalance
	*CommandExecuter
}

func (self *CmdAddBalance) Name() string {
	return self.name
}

func (self *CmdAddBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddBalance) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetBalance{TOR: utils.MONETARY, Overwrite: false}
	}
	return self.rpcParams
}

func (self *CmdAddBalance) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAddBalance) RpcResult() interface{} {
	var s string
	return &s
}
