
package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdRemoveAliases{
		name:      "aliases_remove",
		rpcMethod: "AliasesV1.RemoveAlias",
		rpcParams: &engine.Alias{Direction: utils.OUT},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveAliases struct {
	name      string
	rpcMethod string
	rpcParams *engine.Alias
	*CommandExecuter
}

func (self *CmdRemoveAliases) Name() string {
	return self.name
}

func (self *CmdRemoveAliases) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveAliases) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.Alias{Direction: utils.OUT}
	}
	return self.rpcParams
}

func (self *CmdRemoveAliases) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveAliases) RpcResult() interface{} {
	var s string
	return &s
}
