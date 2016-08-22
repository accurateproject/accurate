package console

import "github.com/accurateproject/accurate/engine"

func init() {
	c := &CmdSetUser{
		name:      "user_set",
		rpcMethod: "UsersV1.SetUser",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdSetUser struct {
	name      string
	rpcMethod string
	rpcParams *engine.UserProfile
	*CommandExecuter
}

func (self *CmdSetUser) Name() string {
	return self.name
}

func (self *CmdSetUser) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetUser) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.UserProfile{}
	}
	return self.rpcParams
}

func (self *CmdSetUser) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSetUser) RpcResult() interface{} {
	var s string
	return &s
}
