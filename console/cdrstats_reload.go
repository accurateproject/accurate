package console

import (
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdCdrReloadQueues{
		name:      "cdrstats_reload",
		rpcMethod: "CDRStatsV1.ReloadQueues",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCdrReloadQueues struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrStatsQueueIDs
	*CommandExecuter
}

func (self *CmdCdrReloadQueues) Name() string {
	return self.name
}

func (self *CmdCdrReloadQueues) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdrReloadQueues) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrStatsQueueIDs{}
	}
	return self.rpcParams
}

func (self *CmdCdrReloadQueues) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdrReloadQueues) RpcResult() interface{} {
	var s string
	return &s
}
