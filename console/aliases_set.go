package console

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdSetAliases{
		name:      "aliases_set",
		rpcMethod: "AliasesV1.SetAlias",
		rpcParams: &engine.AttrAddAlias{Alias: &engine.Alias{Direction: utils.OUT}},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetAliases struct {
	name      string
	rpcMethod string
	rpcParams *engine.AttrAddAlias
	*CommandExecuter
}

func (self *CmdSetAliases) Name() string {
	return self.name
}

func (self *CmdSetAliases) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetAliases) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.AttrAddAlias{Alias: &engine.Alias{Direction: utils.OUT}}
	}
	return self.rpcParams
}

func (self *CmdSetAliases) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetAliases) RpcResult() interface{} {
	var s string
	return &s
}
