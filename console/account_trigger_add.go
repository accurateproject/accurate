package console

import "github.com/accurateproject/accurate/apier/v1"

func init() {
	c := &CmdAccountAddTriggers{
		name:      "account_triggers_add",
		rpcMethod: "ApierV1.AddAccountActionTriggers",
		rpcParams: &v1.AttrAddAccountActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAccountAddTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrAddAccountActionTriggers
	*CommandExecuter
}

func (self *CmdAccountAddTriggers) Name() string {
	return self.name
}

func (self *CmdAccountAddTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAccountAddTriggers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrAddAccountActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdAccountAddTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAccountAddTriggers) RpcResult() interface{} {
	var s string
	return &s
}
