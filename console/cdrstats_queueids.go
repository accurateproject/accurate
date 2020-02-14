package console

import "github.com/accurateproject/accurate/api/v1"

func init() {
	c := &CmdCdrStatsQueueIds{
		name:      "cdrstats_queueids",
		rpcMethod: "CDRStatsV1.GetQueueIDs",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdCdrStatsQueueIds struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrTenant
	*CommandExecuter
}

func (self *CmdCdrStatsQueueIds) Name() string {
	return self.name
}

func (self *CmdCdrStatsQueueIds) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdrStatsQueueIds) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrTenant{}
	}
	return self.rpcParams
}

func (self *CmdCdrStatsQueueIds) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdrStatsQueueIds) RpcResult() interface{} {
	var s []string
	return &s
}

func (self *CmdCdrStatsQueueIds) ClientArgs() (args []string) {
	return
}
