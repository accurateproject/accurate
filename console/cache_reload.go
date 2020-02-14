package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &CmdReloadCache{
		name:      "cache_reload",
		rpcMethod: "ApiV1.ReloadCache",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdReloadCache struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrReloadCache
	rpcResult string
	*CommandExecuter
}

func (self *CmdReloadCache) Name() string {
	return self.name
}

func (self *CmdReloadCache) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdReloadCache) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrReloadCache{}
	}
	return self.rpcParams
}

func (self *CmdReloadCache) PostprocessRpcParams() error {
	return nil
}

func (self *CmdReloadCache) RpcResult() interface{} {
	var s string
	return &s
}
