package console

import "github.com/accurateproject/accurate/engine"

func init() {
	c := &CmdCdrStatsQueue{
		name:      "cdrstats_queue",
		rpcMethod: "CDRStatsV1.GetQueue",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdCdrStatsQueue struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdCdrStatsQueue) Name() string {
	return self.name
}

func (self *CmdCdrStatsQueue) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdrStatsQueue) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdCdrStatsQueue) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdrStatsQueue) RpcResult() interface{} {
	return &engine.StatsQueue{}
}
