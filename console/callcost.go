package console

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdGetCostDetails{
		name:      "cost_details",
		rpcMethod: "ApiV1.GetCallCostLog",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCostDetails struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrGetCallCost
	rpcResult string
	*CommandExecuter
}

func (self *CmdGetCostDetails) Name() string {
	return self.name
}

func (self *CmdGetCostDetails) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCostDetails) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrGetCallCost{RunID: utils.DEFAULT_RUNID}
	}
	return self.rpcParams
}

func (self *CmdGetCostDetails) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetCostDetails) RpcResult() interface{} {
	return &engine.CallCost{}
}
