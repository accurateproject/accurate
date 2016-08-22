package console

import (
	"github.com/accurateproject/accurate/apier/v1"
	"github.com/accurateproject/accurate/engine"
)

func init() {
	c := &CmdGetActionPlan{
		name:      "actionplan_get",
		rpcMethod: "ApierV1.GetActionPlan",
		rpcParams: &v1.AttrGetActionPlan{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetActionPlan struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrGetActionPlan
	*CommandExecuter
}

func (self *CmdGetActionPlan) Name() string {
	return self.name
}

func (self *CmdGetActionPlan) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetActionPlan) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrGetActionPlan{}
	}
	return self.rpcParams
}

func (self *CmdGetActionPlan) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetActionPlan) RpcResult() interface{} {
	s := make([]*engine.ActionPlan, 0)
	return &s
}
