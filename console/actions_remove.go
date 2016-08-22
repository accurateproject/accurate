
package console

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdRemoveActions{
		name:      "actions_remove",
		rpcMethod: "ApierV1.RemoveActions",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdRemoveActions struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrRemoveActions
	*CommandExecuter
}

func (self *CmdRemoveActions) Name() string {
	return self.name
}

func (self *CmdRemoveActions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdRemoveActions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrRemoveActions{}
	}
	return self.rpcParams
}

func (self *CmdRemoveActions) PostprocessRpcParams() error {
	return nil
}

func (self *CmdRemoveActions) RpcResult() interface{} {
	var s string
	return &s
}
