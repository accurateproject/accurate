
package console

func init() {
	c := &CmdPublish{
		name:      "publish",
		rpcMethod: "PubSubV1.Publish",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdPublish struct {
	name      string
	rpcMethod string
	rpcParams *StringMapWrapper
	*CommandExecuter
}

func (self *CmdPublish) Name() string {
	return self.name
}

func (self *CmdPublish) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdPublish) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &StringMapWrapper{}
	}
	return self.rpcParams
}

func (self *CmdPublish) PostprocessRpcParams() error {
	return nil
}

func (self *CmdPublish) RpcResult() interface{} {
	var s string
	return &s
}
