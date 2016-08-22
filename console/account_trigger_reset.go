package console

import "github.com/accurateproject/accurate/apier/v1"

func init() {
	c := &CmdAccountResetTriggers{
		name:      "account_triggers_reset",
		rpcMethod: "ApierV1.ResetAccountActionTriggers",
		rpcParams: &v1.AttrRemoveAccountActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAccountResetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveAccountActionTriggers
	*CommandExecuter
}

func (self *CmdAccountResetTriggers) Name() string {
	return self.name
}

func (self *CmdAccountResetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAccountResetTriggers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveAccountActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdAccountResetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAccountResetTriggers) RpcResult() interface{} {
	var s string
	return &s
}
