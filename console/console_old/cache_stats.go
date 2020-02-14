package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &CmdGetCacheStats{
		name:      "cache_stats",
		rpcMethod: "ApiV1.GetCacheStats",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCacheStats struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrCacheStats
	*CommandExecuter
}

func (self *CmdGetCacheStats) Name() string {
	return self.name
}

func (self *CmdGetCacheStats) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCacheStats) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrCacheStats{}
	}
	return self.rpcParams
}

func (self *CmdGetCacheStats) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetCacheStats) RpcResult() interface{} {
	return &utils.CacheStats{}
}
