
package console

import "github.com/cgrates/cgrates/apier/v2"

func init() {
	c := &CmdAddAccount{
		name:      "account_set",
		rpcMethod: "ApierV2.SetAccount",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAddAccount struct {
	name      string
	rpcMethod string
	rpcParams *v2.AttrSetAccount
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
		self.rpcParams = &v2.AttrSetAccount{}
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
