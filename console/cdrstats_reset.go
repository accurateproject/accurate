
package console

import "github.com/cgrates/cgrates/utils"

func init() {
	c := &CmdCdrResetQueues{
		name:      "cdrstats_reset",
		rpcMethod: "CDRStatsV1.ResetQueues",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCdrResetQueues struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrCDRStatsReloadQueues
	*CommandExecuter
}

func (self *CmdCdrResetQueues) Name() string {
	return self.name
}

func (self *CmdCdrResetQueues) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdrResetQueues) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrCDRStatsReloadQueues{}
	}
	return self.rpcParams
}

func (self *CmdCdrResetQueues) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdrResetQueues) RpcResult() interface{} {
	var s string
	return &s
}
