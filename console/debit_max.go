
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdMaxDebit{
		name:       "debit_max",
		rpcMethod:  "Responder.MaxDebit",
		clientArgs: []string{"Direction", "Category", "TOR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdMaxDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptor
	clientArgs []string
	*CommandExecuter
}

func (self *CmdMaxDebit) Name() string {
	return self.name
}

func (self *CmdMaxDebit) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdMaxDebit) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptor{Direction: "*out"}
	}
	return self.rpcParams
}

func (self *CmdMaxDebit) PostprocessRpcParams() error {
	return nil
}

func (self *CmdMaxDebit) RpcResult() interface{} {
	return &engine.CallCost{}
}

func (self *CmdMaxDebit) ClientArgs() []string {
	return self.clientArgs
}
