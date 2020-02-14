package console

import (
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdExecuteAction{
		name:      "action_execute",
		rpcMethod: "ApiV1.ExecuteAction",
		rpcParams: &utils.AttrExecuteAction{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdExecuteAction struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrExecuteAction
	*CommandExecuter
}

func (self *CmdExecuteAction) Name() string {
	return self.name
}

func (self *CmdExecuteAction) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdExecuteAction) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrExecuteAction{}
	}
	return self.rpcParams
}

func (self *CmdExecuteAction) PostprocessRpcParams() error {
	return nil
}

func (self *CmdExecuteAction) RpcResult() interface{} {
	var s string
	return &s
}
