
package console

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdAccountRemoveTriggers{
		name:      "account_triggers_remove",
		rpcMethod: "ApierV1.RemoveAccountActionTriggers",
		rpcParams: &v1.AttrRemoveAccountActionTriggers{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAccountRemoveTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveAccountActionTriggers
	*CommandExecuter
}

func (self *CmdAccountRemoveTriggers) Name() string {
	return self.name
}

func (self *CmdAccountRemoveTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAccountRemoveTriggers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveAccountActionTriggers{}
	}
	return self.rpcParams
}

func (self *CmdAccountRemoveTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdAccountRemoveTriggers) RpcResult() interface{} {
	var s string
	return &s
}
