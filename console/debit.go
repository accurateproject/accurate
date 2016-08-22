
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdDebit{
		name:       "debit",
		rpcMethod:  "Responder.Debit",
		clientArgs: []string{"Direction", "Category", "TOR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject", "DryRun"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdDebit struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptor
	clientArgs []string
	*CommandExecuter
}

func (self *CmdDebit) Name() string {
	return self.name
}

func (self *CmdDebit) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdDebit) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptor{Direction: "*out"}
	}
	return self.rpcParams
}

func (self *CmdDebit) PostprocessRpcParams() error {
	return nil
}

func (self *CmdDebit) RpcResult() interface{} {
	return &engine.CallCost{}
}

func (self *CmdDebit) ClientArgs() []string {
	return self.clientArgs
}
