
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdUnsubscribe{
		name:      "unsubscribe",
		rpcMethod: "PubSubV1.Unsubscribe",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdUnsubscribe struct {
	name      string
	rpcMethod string
	rpcParams *engine.SubscribeInfo
	*CommandExecuter
}

func (self *CmdUnsubscribe) Name() string {
	return self.name
}

func (self *CmdUnsubscribe) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdUnsubscribe) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.SubscribeInfo{}
	}
	return self.rpcParams
}

func (self *CmdUnsubscribe) PostprocessRpcParams() error {
	return nil
}

func (self *CmdUnsubscribe) RpcResult() interface{} {
	var s string
	return &s
}
