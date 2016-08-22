package console

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdGetMatchingAliases{
		name:      "aliases_matching",
		rpcMethod: "AliasesV1.GetMatchingAlias",
		rpcParams: &engine.AttrMatchingAlias{Direction: utils.OUT},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMatchingAliases struct {
	name      string
	rpcMethod string
	rpcParams *engine.AttrMatchingAlias
	*CommandExecuter
}

func (self *CmdGetMatchingAliases) Name() string {
	return self.name
}

func (self *CmdGetMatchingAliases) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMatchingAliases) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.AttrMatchingAlias{Direction: utils.OUT}
	}
	return self.rpcParams
}

func (self *CmdGetMatchingAliases) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMatchingAliases) RpcResult() interface{} {
	var s string
	return &s
}
