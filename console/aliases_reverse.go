
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdGetReverseAliases{
		name:      "aliases_reverse",
		rpcMethod: "AliasesV1.GetReverseAlias",
		rpcParams: &engine.AttrReverseAlias{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetReverseAliases struct {
	name      string
	rpcMethod string
	rpcParams *engine.AttrReverseAlias
	*CommandExecuter
}

func (self *CmdGetReverseAliases) Name() string {
	return self.name
}

func (self *CmdGetReverseAliases) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetReverseAliases) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.AttrReverseAlias{}
	}
	return self.rpcParams
}

func (self *CmdGetReverseAliases) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetReverseAliases) RpcResult() interface{} {
	a := make(map[string][]*engine.Alias)
	return &a
}
