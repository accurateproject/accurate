package console

import "github.com/accurateproject/accurate/apier/v1"

func init() {
	c := &CmdSetTriggers{
		name:      "triggers_set",
		rpcMethod: "ApierV1.SetActionTrigger",
		rpcParams: &v1.AttrSetActionTrigger{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetTriggers struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetActionTrigger
	*CommandExecuter
}

func (self *CmdSetTriggers) Name() string {
	return self.name
}

func (self *CmdSetTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetTriggers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetActionTrigger{}
	}
	return self.rpcParams
}

func (self *CmdSetTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetTriggers) RpcResult() interface{} {
	var s string
	return &s
}
