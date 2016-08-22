
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdGetMaxUsage{
		name:      "maxusage",
		rpcMethod: "ApierV1.GetMaxUsage",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxUsage struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.UsageRecord
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxUsage) Name() string {
	return self.name
}

func (self *CmdGetMaxUsage) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxUsage) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.UsageRecord{Direction: "*out"}
	}
	return self.rpcParams
}

func (self *CmdGetMaxUsage) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxUsage) RpcResult() interface{} {
	var f float64
	return &f
}

func (self *CmdGetMaxUsage) ClientArgs() []string {
	return self.clientArgs
}
