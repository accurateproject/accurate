package console

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func init() {
	c := &CmdGetAccounts{
		name:      "accounts",
		rpcMethod: "ApierV2.GetAccounts",
		rpcParams: &utils.AttrGetAccounts{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetAccounts struct {
	name      string
	rpcMethod string
	rpcParams *utils.AttrGetAccounts
	*CommandExecuter
}

func (self *CmdGetAccounts) Name() string {
	return self.name
}

func (self *CmdGetAccounts) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetAccounts) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.AttrGetAccounts{}
	}
	return self.rpcParams
}

func (self *CmdGetAccounts) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetAccounts) RpcResult() interface{} {
	a := make([]engine.Account, 0)
	return &a
}
