package console

import "github.com/accurateproject/accurate/engine"

func init() {
	c := &CmdGetCost{
		name:       "cost",
		rpcMethod:  "Responder.GetCost",
		clientArgs: []string{"Direction", "Category", "TOR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetCost struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptor
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetCost) Name() string {
	return self.name
}

func (self *CmdGetCost) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetCost) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptor{Direction: "*out"}
	}
	return self.rpcParams
}

func (self *CmdGetCost) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetCost) RpcResult() interface{} {
	return &engine.CallCost{}
}

func (self *CmdGetCost) ClientArgs() []string {
	return self.clientArgs
}
