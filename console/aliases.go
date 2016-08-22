
package console

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	c := &CmdGetAliases{
		name:      "aliases",
		rpcMethod: "AliasesV1.GetAlias",
		rpcParams: &engine.Alias{Direction: utils.OUT},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAliases struct {
	name      string
	rpcMethod string
	rpcParams *engine.Alias
	*CommandExecuter
}

func (self *CmdGetAliases) Name() string {
	return self.name
}

func (self *CmdGetAliases) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAliases) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.Alias{Direction: utils.OUT}
	}
	return self.rpcParams
}

func (self *CmdGetAliases) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAliases) RpcResult() interface{} {
	a := engine.Alias{}
	return &a
}
