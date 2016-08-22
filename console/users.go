package console

import "github.com/accurateproject/accurate/engine"

func init() {
	c := &CmdGetUsers{
		name:      "users",
		rpcMethod: "UsersV1.GetUsers",
		rpcParams: &engine.UserProfile{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetUsers struct {
	name      string
	rpcMethod string
	rpcParams *engine.UserProfile
	*CommandExecuter
}

func (self *CmdGetUsers) Name() string {
	return self.name
}

func (self *CmdGetUsers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetUsers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.UserProfile{}
	}
	return self.rpcParams
}

func (self *CmdGetUsers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetUsers) RpcResult() interface{} {
	s := engine.UserProfiles{}
	return &s
}
