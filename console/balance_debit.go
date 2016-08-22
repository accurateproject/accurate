
package console

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdBalanceDebit{
		name:      "balance_debit",
		rpcMethod: "ApierV1.DebitBalance",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdBalanceDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *v1.AttrAddBalance
	clientArgs []string
	*CommandExecuter
}

func (self *CmdBalanceDebit) Name() string {
	return self.name
}

func (self *CmdBalanceDebit) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdBalanceDebit) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &v1.AttrAddBalance{}
	}
	return self.rpcParams
}

func (self *CmdBalanceDebit) PostprocessRpcParams() error {
	return nil
}

func (self *CmdBalanceDebit) RpcResult() interface{} {
	var s string
	return &s
}
