package console

import (
	"github.com/accurateproject/accurate/apier/v1"
)

func init() {
	c := &CmdCdrcConfigReload{
		name:      "cdrc_config_reload",
		rpcMethod: "ApierV1.ReloadCdrcConfig",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdCdrcConfigReload struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrReloadConfig
	*CommandExecuter
}

func (self *CmdCdrcConfigReload) Name() string {
	return self.name
}

func (self *CmdCdrcConfigReload) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdCdrcConfigReload) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(v1.AttrReloadConfig)
	}
	return self.rpcParams
}

func (self *CmdCdrcConfigReload) PostprocessRpcParams() error {
	return nil
}

func (self *CmdCdrcConfigReload) RpcResult() interface{} {
	var s string
	return &s
}
