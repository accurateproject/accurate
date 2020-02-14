package console

import (
	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdSetBalance{
		name:      "balance_set",
		rpcMethod: "ApiV1.SetBalance",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetBalance struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetBalance
	*CommandExecuter
}

func (self *CmdSetBalance) Name() string {
	return self.name
}

func (self *CmdSetBalance) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetBalance) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetBalance{TOR: utils.MONETARY}
	}
	return self.rpcParams
}

func (self *CmdSetBalance) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetBalance) RpcResult() interface{} {
	var s string
	return &s
}
