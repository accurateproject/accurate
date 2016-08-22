
package console

func init() {
	c := &CmdReloadScheduler{
		name:      "scheduler_reload",
		rpcMethod: "ApierV1.ReloadScheduler",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdReloadScheduler struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdReloadScheduler) Name() string {
	return self.name
}

func (self *CmdReloadScheduler) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdReloadScheduler) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdReloadScheduler) PostprocessRpcParams() error {
	return nil
}

func (self *CmdReloadScheduler) RpcResult() interface{} {
	var s string
	return &s
}
