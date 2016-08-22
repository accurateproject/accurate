
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdSubscribe{
		name:      "subscribe",
		rpcMethod: "PubSubV1.Subscribe",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdSubscribe struct {
	name      string
	rpcMethod string
	rpcParams *engine.SubscribeInfo
	*CommandExecuter
}

func (self *CmdSubscribe) Name() string {
	return self.name
}

func (self *CmdSubscribe) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSubscribe) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.SubscribeInfo{}
	}
	return self.rpcParams
}

func (self *CmdSubscribe) PostprocessRpcParams() error {
	return nil
}

func (self *CmdSubscribe) RpcResult() interface{} {
	var s string
	return &s
}
