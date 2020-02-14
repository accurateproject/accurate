package console

import "github.com/accurateproject/accurate/apier/v1"

func init() {
	c := &CmdSetActionPlan{
		name:      "actionplan_set",
		rpcMethod: "ApiV1.SetActionPlan",
		rpcParams: &v1.AttrSetActionPlan{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetActionPlan
	*CommandExecuter
}

func (self *CmdSetActionPlan) Name() string {
	return self.name
}

func (self *CmdSetActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetActionPlan) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetActionPlan{}
	}
	return self.rpcParams
}

func (self *CmdSetActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetActionPlan) RpcResult() interface{} {
	var s string
	return &s
}
