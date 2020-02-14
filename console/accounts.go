package console

import (
	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/engine"
)

func init() {
	c := &CmdGetAccounts{
		name:      "accounts",
		rpcMethod: "ApiV1.GetAccounts",
		rpcParams: &v1.AttrGetMultiple{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAccounts struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetMultiple
	*CommandExecuter
}

func (self *CmdGetAccounts) Name() string {
	return self.name
}

func (self *CmdGetAccounts) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAccounts) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetMultiple{}
	}
	return self.rpcParams
}

func (self *CmdGetAccounts) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAccounts) RpcResult() interface{} {
	a := make([]*engine.Account, 0)
	return &a
}
