
package console

func init() {
	c := &CmdUserAddIndex{
		name:      "user_addindex",
		rpcMethod: "UsersV1.AddIndex",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdUserAddIndex struct {
	name      string
	rpcMethod string
	rpcParams *StringSliceWrapper
	*CommandExecuter
}

func (self *CmdUserAddIndex) Name() string {
	return self.name
}

func (self *CmdUserAddIndex) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdUserAddIndex) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringSliceWrapper{}
	}
	return self.rpcParams
}

func (self *CmdUserAddIndex) PostprocessRpcParams() error {
	return nil
}

func (self *CmdUserAddIndex) RpcResult() interface{} {
	var s string
	return &s
}
