
package console

func init() {
	c := &CmdUserShowIndexes{
		name:      "user_indexes",
		rpcMethod: "UsersV1.GetIndexes",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdUserShowIndexes struct {
	name      string
	rpcMethod string
	rpcParams *EmptyWrapper
	*CommandExecuter
}

func (self *CmdUserShowIndexes) Name() string {
	return self.name
}

func (self *CmdUserShowIndexes) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdUserShowIndexes) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &EmptyWrapper{}
	}
	return self.rpcParams
}

func (self *CmdUserShowIndexes) PostprocessRpcParams() error {
	return nil
}

func (self *CmdUserShowIndexes) RpcResult() interface{} {
	s := map[string][]string{}
	return &s
}
