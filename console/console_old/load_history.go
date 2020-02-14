package console

import "github.com/accurateproject/accurate/utils"

func init() {
	c := &CmdGetLoadHistory{
		name:      "load_history",
		rpcMethod: "ApiV1.GetLoadHistory",
		rpcParams: new(utils.Paginator),
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Returns the list of load items from the history, in reverse order
type CmdGetLoadHistory struct {
	name      string
	rpcMethod string
	rpcParams *utils.Paginator
	*CommandExecuter
}

func (self *CmdGetLoadHistory) Name() string {
	return self.name
}

func (self *CmdGetLoadHistory) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetLoadHistory) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = new(utils.Paginator)
	}
	return self.rpcParams
}

func (self *CmdGetLoadHistory) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetLoadHistory) RpcResult() interface{} {
	a := make([]*utils.LoadInstance, 0)
	return &a
}
