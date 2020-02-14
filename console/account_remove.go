package console

import (
	"github.com/accurateproject/accurate/api/v1"
)

func init() {
	c := &CmdRemoveAccount{
		name:      "account_remove",
		rpcMethod: "ApiV1.RemoveAccount",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveAccount struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveAccount
	*CommandExecuter
}

func (self *CmdRemoveAccount) Name() string {
	return self.name
}

func (self *CmdRemoveAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveAccount) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveAccount{}
	}
	return self.rpcParams
}

func (self *CmdRemoveAccount) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveAccount) RpcResult() interface{} {
	var s string
	return &s
}
