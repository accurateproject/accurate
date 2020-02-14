package console

import "github.com/accurateproject/accurate/apier/v1"

func init() {
	c := &CmdRemoveTriggers{
		name:      "triggers_remove",
		rpcMethod: "ApiV1.RemoveActionTrigger",
		rpcParams: &v1.AttrRemoveActionTrigger{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveActionTrigger
	*CommandExecuter
}

func (self *CmdRemoveTriggers) Name() string {
	return self.name
}

func (self *CmdRemoveTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveTriggers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveActionTrigger{}
	}
	return self.rpcParams
}

func (self *CmdRemoveTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveTriggers) RpcResult() interface{} {
	var s string
	return &s
}
