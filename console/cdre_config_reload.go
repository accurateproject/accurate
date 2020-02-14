package console

import (
	"github.com/accurateproject/accurate/api/v1"
)

func init() {
	c := &CmdCdreConfigReload{
		name:      "cdre_config_reload",
		rpcMethod: "ApiV1.ReloadCdreConfig",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCdreConfigReload struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrReloadConfig
	*CommandExecuter
}

func (self *CmdCdreConfigReload) Name() string {
	return self.name
}

func (self *CmdCdreConfigReload) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdreConfigReload) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(v1.AttrReloadConfig)
	}
	return self.rpcParams
}

func (self *CmdCdreConfigReload) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdreConfigReload) RpcResult() interface{} {
	var s string
	return &s
}
