
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdUserRemove{
		name:      "user_remove",
		rpcMethod: "UsersV1.RemoveUser",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdUserRemove struct {
	name      string
	rpcMethod string
	rpcParams *engine.UserProfile
	*CommandExecuter
}

func (self *CmdUserRemove) Name() string {
	return self.name
}

func (self *CmdUserRemove) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdUserRemove) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.UserProfile{}
	}
	return self.rpcParams
}

func (self *CmdUserRemove) PostprocessRpcParams() error {
	return nil
}

func (self *CmdUserRemove) RpcResult() interface{} {
	var s string
	return &s
}
