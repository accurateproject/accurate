
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdCdrStatsQueueTriggers{
		name:      "cdrstats_queue_triggers",
		rpcMethod: "CDRStatsV1.GetQueueTriggers",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdCdrStatsQueueTriggers struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdCdrStatsQueueTriggers) Name() string {
	return self.name
}

func (self *CmdCdrStatsQueueTriggers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdrStatsQueueTriggers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdCdrStatsQueueTriggers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdrStatsQueueTriggers) RpcResult() interface{} {
	return new(engine.ActionTriggers)
}
