
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdGetSharedGroup{
		name:      "sharedgroup",
		rpcMethod: "ApierV1.GetSharedGroup",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetSharedGroup struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdGetSharedGroup) Name() string {
	return self.name
}

func (self *CmdGetSharedGroup) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetSharedGroup) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdGetSharedGroup) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetSharedGroup) RpcResult() interface{} {
	return &engine.SharedGroup{}
}
