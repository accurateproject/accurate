
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdUpdateUser{
		name:      "user_update",
		rpcMethod: "UsersV1.UpdateUser",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdUpdateUser struct {
	name      string
	rpcMethod string
	rpcParams *engine.UserProfile
	*CommandExecuter
}

func (self *CmdUpdateUser) Name() string {
	return self.name
}

func (self *CmdUpdateUser) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdUpdateUser) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.UserProfile{}
	}
	return self.rpcParams
}

func (self *CmdUpdateUser) PostprocessRpcParams() error {
	return nil
}

func (self *CmdUpdateUser) RpcResult() interface{} {
	var s string
	return &s
}
