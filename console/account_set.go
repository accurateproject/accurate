package console

import "github.com/accurateproject/accurate/api/v1"

func init() {
	c := &CmdAddAccount{
		name:      "account_set",
		rpcMethod: "ApiV1.SetAccount",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAddAccount struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetAccount
	*CommandExecuter
}

func (self *CmdAddAccount) Name() string {
	return self.name
}

func (self *CmdAddAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddAccount) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetAccount{}
	}
	return self.rpcParams
}

func (self *CmdAddAccount) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAddAccount) RpcResult() interface{} {
	var s string
	return &s
}
