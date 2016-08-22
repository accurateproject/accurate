
package console

import "github.com/cgrates/cgrates/engine"

func init() {
	c := &CmdShowSubscribers{
		name:      "subscribers",
		rpcMethod: "PubSubV1.ShowSubscribers",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdShowSubscribers struct {
	name      string
	rpcMethod string
	rpcParams *StringWrapper
	*CommandExecuter
}

func (self *CmdShowSubscribers) Name() string {
	return self.name
}

func (self *CmdShowSubscribers) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdShowSubscribers) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringWrapper{}
	}
	return self.rpcParams
}

func (self *CmdShowSubscribers) PostprocessRpcParams() error {
	return nil
}

func (self *CmdShowSubscribers) RpcResult() interface{} {
	var s map[string]*engine.SubscriberData
	return &s
}
